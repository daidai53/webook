// Copyright@daidai53 2023
package middleware

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type LogMiddlewareBuilder struct {
	logFn         func(ctx context.Context, l AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLogMiddlewareBuilder(logFn func(ctx context.Context, l AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		logFn: logFn,
	}
}

func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	l.allowReqBody = true
	return l
}

func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	l.allowRespBody = true
	return l
}

func (l *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) > 1024 {
			path = path[:1024]
		}
		method := c.Request.Method
		al := AccessLog{
			Path:   path,
			Method: method,
		}
		if l.allowReqBody {
			body, _ := c.GetRawData()
			if len(body) > 2048 {
				al.ReqBody = string(body[:2048])
			} else {
				al.ReqBody = string(body)
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}

		start := time.Now()

		if l.allowRespBody {
			c.Writer = &responseWriter{
				al:             &al,
				ResponseWriter: c.Writer,
			}
		}

		defer func() {
			duration := time.Since(start)
			al.Duration = duration
			l.logFn(c, al)
		}()

		// 直接执行下一个middleware
		c.Next()
	}
}

type AccessLog struct {
	Path       string        `json:"path"`
	Method     string        `json:"method"`
	ReqBody    string        `json:"reqBody"`
	RespBody   string        `json:"respBody"`
	StatusCode int           `json:"statusCode"`
	Duration   time.Duration `json:"duration"`
}

type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.al.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
