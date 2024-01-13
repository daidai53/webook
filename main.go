// Copyright@daidai53 2023
package main

import (
	"context"
	"github.com/daidai53/webook/ioc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"net/http"
	"time"
)

func main() {
	initViperRemote()
	initPrometheus()
	tpCancel := ioc.InitOTEL()
	defer func() {
		tpCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		tpCancel(tpCtx)
	}()
	app := InitWebServer()
	server := app.server
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	server.Run(":8081")
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3",
		"http://localhost:12379", "D:/Program Files/Git/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8082", nil)
	}()
}
