// Copyright@daidai53 2024
package k6

import (
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestHello(t *testing.T) {
	server := gin.Default()
	server.POST("/hello", func(context *gin.Context) {
		var u User
		if err := context.Bind(&u); err != nil {
			return
		}
		number := rand.Int31n(1000)
		time.Sleep(time.Millisecond * time.Duration(number))
		if number%100 < 10 {
			context.String(http.StatusInternalServerError, "模拟服务器失败")
		} else {
			context.String(http.StatusOK, u.Name)
		}
	})
	server.Run(":8080")
}

type User struct {
	Name string `json:"name"`
}
