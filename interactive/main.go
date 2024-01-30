// Copyright@daidai53 2023
package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log"
	"net/http"
)

func main() {
	initViper()
	initPrometheus()
	//tpCancel := ioc.InitOTEL()
	//defer func() {
	//	tpCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	//	defer cancel()
	//	tpCancel(tpCtx)
	//}()
	app := InitApp()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	go func() {
		err1 := app.adminServer.Start()
		panic(err1)
	}()
	app.server.Serve()
}

func initViper() {
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	// 当前工作目录的 config 子目录
	viper.AddConfigPath("config")
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test.key")
	log.Println(val)
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8082", nil)
	}()
}
