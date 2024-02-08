// Copyright@daidai53 2024
package wrr

import (
	"context"
	"errors"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

const Name = "custom_weighted_round_robin"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &PickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	for sc, sci := range info.ReadySCs {
		md, _ := sci.Address.Metadata.(map[string]any)
		weightVal, _ := md["weight"]
		weight, _ := weightVal.(float64)
		if weight == 0 {
			weight = 1
		}
		conns = append(conns, &weightConn{
			SubConn:       sc,
			weight:        int(weight),
			currentWeight: int(weight),
			available:     true,
		})
	}

	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	conns []*weightConn
	lock  sync.Mutex
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var total int
	var maxCC *weightConn
	for _, c := range p.conns {
		if !c.available {
			continue
		}
		total += c.weight
		c.currentWeight += c.weight
		if maxCC == nil || maxCC.currentWeight < c.currentWeight {
			maxCC = c
		}
	}

	maxCC.currentWeight -= total

	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			p.adjustWeightV2(info, maxCC)
		},
	}, nil
}

func (p *Picker) adjustWeight(info balancer.DoneInfo, selected *weightConn) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if info.Err == nil {
		var total int
		for _, c := range p.conns {
			if !c.available {
				continue
			}
			total += c.weight
		}
		selected.currentWeight += len(p.conns) * selected.weight
		if selected.currentWeight > total {
			selected.weight = total
		}
		return
	}
	selected.currentWeight -= len(p.conns) * selected.weight
	if selected.currentWeight < 0 {
		selected.currentWeight = 0
	}
}

func (p *Picker) adjustWeightV2(info balancer.DoneInfo, selected *weightConn) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// 返回error为空时，适当调高当前所选节点的权重
	if info.Err == nil {
		var total int
		for _, c := range p.conns {
			if !c.available {
				continue
			}
			total += c.weight
		}
		selected.currentWeight += len(p.conns) * selected.weight
		if selected.currentWeight > total {
			selected.weight = total
		}
		return
	}
	status, ok := status.FromError(info.Err)
	if !ok {
		// 从error中获取grpc调用结果失败
		return
	}
	switch status.Code() {
	case codes.Unavailable:
		selected.available = false
		go func() {
			ticker := time.NewTicker(time.Minute * 5)
			ctx, cancel := context.WithTimeout(context.Background(), time.Hour*24)
			defer cancel()

			for {
				select {
				case <-ticker.C:
					if err1 := p.connectPeer(); err1 == nil {
						p.lock.Lock()
						selected.available = true
						p.lock.Unlock()
					}
				case <-ctx.Done():
					// 超时未连接上
					return
				default:
					return
				}
			}
		}()
	case codes.ResourceExhausted:
		selected.currentWeight -= len(p.conns) * selected.weight
		if selected.currentWeight < 0 {
			selected.currentWeight = 0
		}
	default:
		selected.currentWeight -= selected.weight
		if selected.currentWeight < 0 {
			selected.currentWeight = 0
		}
	}
}

// 向服务端发送健康检查请求
func (p *Picker) connectPeer() error {
	return errors.New("未实现")
}

type weightConn struct {
	balancer.SubConn
	weight        int
	currentWeight int
	available     bool
}
