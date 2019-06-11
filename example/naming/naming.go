package main

import (
	"gitlab.mfwdev.com/portal/go-nacos"
	"gitlab.mfwdev.com/portal/go-nacos/api"
	"gitlab.mfwdev.com/portal/go-nacos/pkg/util"
	"gitlab.mfwdev.com/portal/go-nacos/types"
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
	ns := app.NewNamingService()
	er := ns.RegisterInstance(&types.ServiceInstance{
		GroupName:   "beta",
		ServiceName: "local-2",
		IP:          "10.10.10.15",
		Port:        8080,
		Metadata: util.MapToString(map[string]string{
			"name": "go-nacos",
			"age":  "30",
		}),
		Weight:    1.0,
		Healthy:   true,
		Enable:    true,
		Ephemeral: false,
	})
	if er != nil {
		panic(er)
	}
}
