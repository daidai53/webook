// Copyright@daidai53 2023
package main

import (
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func main() {
	initViperRemote()
	server := InitWebServer()
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
