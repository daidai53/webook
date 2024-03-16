// Copyright@daidai53 2024
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"sync"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "C:\\ProgramData\\elasticsearch\\logs\\user.log",
		MaxSize:    2,
		MaxBackups: 100,
		MaxAge:     28,
		Compress:   true,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(lumberjackLogger),
		zapcore.DebugLevel,
	)

	l := zap.New(core, zap.AddCaller())
	res := NewZapLogger(l)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// 为了演示 ELK，我直接输出日志
		ticker := time.NewTicker(time.Nanosecond * 1)
		for t := range ticker.C {
			res.Info("模拟输出日志", String("time", t.String()))
		}
		wg.Done()
	}()
	wg.Wait()
}
