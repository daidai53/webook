// Copyright@daidai53 2024
package cache

import (
	"container/heap"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/daidai53/webook/interactive/repository/dao"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"strconv"
	"sync"
	"time"
)

//go:embed lua/top_incr_like.lua
var topIncrLikes string

type TopLikesArticleCache interface {
	Init(ctx context.Context, likes []dao.Likes)
	GetTopLikesIds(ctx context.Context, n int) ([]int64, error)
	IncrLike(ctx context.Context, id int64, delta int64) error
}

type topLikesCache struct {
	init  bool
	cmd   redis.Cmdable
	likes likes
	top10 Top10Likes
	l     logger.LoggerV1
}

func NewTopLikesCache(cmd redis.Cmdable, l logger.LoggerV1) TopLikesArticleCache {
	ret := &topLikesCache{
		cmd: cmd,
		l:   l,
	}
	ret.likes.mu.Lock()
	ret.likes.Likes = make(Likes, 0)
	heap.Init(&ret.likes.Likes)
	ret.likes.mu.Unlock()
	return ret
}

func (t *topLikesCache) Init(ctx context.Context, likes []dao.Likes) {
	for _, like := range likes {
		t.likes.mu.Lock()
		heap.Push(&t.likes.Likes, Like{
			artiId:  like.BizId,
			likeCnt: like.LikeCnt,
		})
		t.likes.mu.Unlock()
		err := t.cmd.ZAdd(ctx, "article:likes:top", redis.Z{
			Score:  float64(like.LikeCnt),
			Member: fmt.Sprintf("%d", like.BizId),
		}).Err()
		if err != nil {
			t.l.Error("初始化时插入redis失败",
				logger.Error(err),
				logger.Int64("aid", like.BizId),
				logger.Int64("like_cnt", like.LikeCnt))
		}
	}
	for i := 1; i <= 10; i++ {
		tmp, err := t.topLikesHeap(i)
		if err != nil {
			t.l.Debug("从堆中获取top结果失败", logger.Error(err))
			continue
		}
		t.top10.update(i, tmp)
	}
	go func() {

	}()
	t.init = true
}

func (t *topLikesCache) Refresh() {
	for {
		time.Sleep(time.Second * 30)
		t.top10.mu.Lock()
		for i := 1; i <= 10; i++ {
			tmp, err := t.GetTopLikesIds(nil, i)
			if err != nil {
				t.l.Error("刷新时获取数据失败",
					logger.Error(err),
					logger.Int("n", i))
			}
			t.top10.update(i, tmp)
		}
		t.top10.mu.Unlock()
	}
}

func (t *topLikesCache) IncrLike(ctx context.Context, id int64, delta int64) error {
	if !t.init {
		return errors.New("not init")
	}
	var eg errgroup.Group
	eg.Go(func() error {
		t.incrLike(id)
		return nil
	})
	eg.Go(func() error {
		return t.cmd.ZIncrBy(ctx, "article:likes:top", float64(delta), fmt.Sprintf("%d", id)).Err()
	})
	return eg.Wait()
}

func (t *topLikesCache) GetTopLikesIds(ctx context.Context, n int) ([]int64, error) {
	if !t.init {
		return nil, errors.New("not init")
	}
	var res []int64
	if n <= 10 {
		res, err := t.top10.GetTop(n)
		if err == nil {
			return res, nil
		}
		t.l.Error("从top10缓存中没有读到结果",
			logger.Error(err))
	}
	resCh := make(chan []int64)
	sent := false
	ctxNew, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() {
		heapRes, err := t.topLikesHeap(n)
		if err != nil {
			// log
			return
		}
		if !sent {
			sent = true
			resCh <- heapRes
		}
	}()
	go func() {
		redisRes, err := t.cmd.ZRevRangeByScoreWithScores(ctxNew, "article:likes:top", &redis.ZRangeBy{
			Min:   "-inf",
			Max:   "+inf",
			Count: int64(n),
		}).Result()
		if err != nil {
			// log
			return
		}
		res = slice.Map[redis.Z, int64](redisRes, func(idx int, src redis.Z) int64 {
			i, _ := strconv.ParseInt(src.Member.(string), 10, 64)
			return i
		})
		if !sent {
			sent = true
			resCh <- res
		}
	}()

	select {
	case ids := <-resCh:
		close(resCh)
		go func() {
			if n <= 10 && n > 0 {
				t.top10.update(n, ids)
			}
		}()
		return ids, nil
	case <-ctxNew.Done():
		sent = true
		close(resCh)
		return nil, errors.New("超时未返回top结果")
	}
}

func (t *topLikesCache) incrLike(id int64) {
	t.likes.mu.Lock()
	defer t.likes.mu.Unlock()
	for i, like := range t.likes.Likes {
		if like.artiId == id {
			t.likes.Likes[i].likeCnt++
			return
		}
	}
	heap.Push(&t.likes.Likes, Like{
		artiId:  id,
		likeCnt: 1,
	})
}

func (t *topLikesCache) topLikesHeap(n int) ([]int64, error) {
	res := make([]int64, 0, n)
	popOuts := make([]Like, 0, n)
	var tmp Like
	var ok bool
	t.likes.mu.Lock()
	defer t.likes.mu.Unlock()
	defer func() {
		for _, like := range popOuts {
			heap.Push(&t.likes.Likes, like)
		}
	}()
	if len(t.likes.Likes) == 0 {
		return nil, errors.New("堆空")
	}
	for i := 0; i < n; i++ {
		if len(t.likes.Likes) == 0 {
			return res, nil
		}
		tmp, ok = heap.Pop(&t.likes.Likes).(Like)
		if !ok {
			return nil, errors.New("堆中存储非int64元素")
		}
		res = append(res, tmp.artiId)
		popOuts = append(popOuts, tmp)
	}
	return res, nil
}

type Top10Likes struct {
	mu    sync.RWMutex
	top10 [10]top10Like
}

func (t *Top10Likes) GetTop(n int) ([]int64, error) {
	if n <= 0 || n > 10 {
		return nil, errors.New("invalid input")
	}
	n = n - 1
	if !t.top10[n].valid {
		return nil, errors.New("在Top10缓存中没有找到")
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.top10[n].ids, nil
}

func (t *Top10Likes) update(n int, value []int64) {
	n = n - 1
	t.mu.Lock()
	t.top10[n].ids = value
	t.top10[n].valid = true
	t.mu.Unlock()
}

type top10Like struct {
	valid bool
	ids   []int64
}

type Like struct {
	likeCnt int64
	artiId  int64
}

type Likes []Like

type likes struct {
	Likes
	mu sync.Mutex
}

func (l *Likes) Len() int {
	return len(*l)
}

func (l *Likes) Less(i, j int) bool {
	if (*l)[i].likeCnt > (*l)[j].likeCnt {
		return true
	}
	return false
}

func (l *Likes) Swap(i, j int) {
	(*l)[i], (*l)[j] = (*l)[j], (*l)[i]
}

func (l *Likes) Push(x any) {
	likeElem, ok := x.(Like)
	if ok {
		*l = append(*l, likeElem)
	}
}

func (l *Likes) Pop() any {
	ret := (*l)[len(*l)-1]
	*l = (*l)[0 : len(*l)-1]
	return ret
}
