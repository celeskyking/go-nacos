package main

import (
	"fmt"
	"github.com/celeskyking/go-nacos"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/config"
	"github.com/celeskyking/go-nacos/pkg/util"
)

//
//name=go-nacos
//size=20
//
//
//
func main() {
	appConfig := &api.AppConfig{
		AppName: "demo",
		Env:     "dev",
		Port:    8080,
		//IP可以为空
		IP: util.LocalIP(),
	}
	app := nacos.NewApplication(appConfig)
	configService := app.NewConfigService("/tmp/nacos/config/snapshot")
	//目前只支持properties文件,不过支持自定义格式文件的扩展,Custom方法
	file, er := configService.Properties("demo.properties")
	if er != nil {
		panic(er)
	}
	file.ListenValue("name", func(key string, curValue, newValue string, ctx *config.FileDesc) {
		fmt.Println("new value:" + newValue)
	})
	n := file.MustGet("name")
	fmt.Printf("name:%s", n)
	s := file.MustGetInt32("size")
	fmt.Printf("size:%d\n", s)
	configService.StopWatch()
}
