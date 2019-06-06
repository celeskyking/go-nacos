package main

import (
	"fmt"
	"github.com/celeskyking/go-nacos"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/naming"
	"github.com/celeskyking/go-nacos/pkg/util"
)

func main() {
	appConfig := &api.AppConfig{
		AppName: "demo",
		Env:     "dev",
		Port:    8080,
		//IP可以为空
		IP: util.LocalIP(),
	}
	app := nacos.NewApplication(appConfig)
	dc := app.NewDiscoveryClient()
	er := dc.Register()
	if er != nil {
		panic(er)
	}
	serverList, er := dc.GetInstances("app2", &naming.QueryOptions{
		Namespace: "",
		Cluster:   "",
		//会接受推送
		Watch: true,
	})
	if er != nil {
		panic(er)
	}
	instance := serverList.SelectOne()
	fmt.Println(util.ToJSONString(instance))
	all := serverList.GetAll()
	fmt.Println(util.ToJSONString(all))
	serverList.StopListen()
	er = dc.Deregister()
	if er != nil {
		panic(er)
	}
}
