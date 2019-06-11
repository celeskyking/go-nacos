package main

import (
	"fmt"
	"gitlab.mfwdev.com/portal/go-nacos"
	"gitlab.mfwdev.com/portal/go-nacos/api"
	"gitlab.mfwdev.com/portal/go-nacos/naming"
	"gitlab.mfwdev.com/portal/go-nacos/pkg/util"
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
	app.SetServers(&api.ServerOptions{
		Addresses:       []string{"127.0.0.1:8848"},
		LBStrategy:      api.RoundRobin,
		EndpointEnabled: false,
	})
	dc := app.NewDiscoveryClient()
	er := dc.Register()
	if er != nil {
		panic(er)
	}
	serverList, er := dc.GetInstances("demo", &naming.QueryOptions{
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
