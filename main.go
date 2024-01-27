// Copyright@daidai53 2023
package main

import (
	"context"
	"github.com/daidai53/webook/ioc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log"
	"net/http"
	"time"
)

func main() {
	initViperWatch()
	initPrometheus()
	tpCancel := ioc.InitOTEL()
	defer func() {
		tpCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		tpCancel(tpCtx)
	}()
	app := InitWebServer()
	app.cron.Start()
	defer func() {
		<-app.cron.Stop().Done()
	}()
	server := app.server
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	server.Run(":8081")
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

func initViperWatch() {
	cfile := pflag.String("config",
		"config/dev.yaml", "webook")
	pflag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test.key")
	log.Println(val)
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
