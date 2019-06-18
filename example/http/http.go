package main

import (
	"fmt"
	"gitlab.mfwdev.com/portal/go-nacos"
	"gitlab.mfwdev.com/portal/go-nacos/api"
	"gitlab.mfwdev.com/portal/go-nacos/pkg/util"
	"gitlab.mfwdev.com/portal/go-nacos/types"
)

func main() {
	appConfig := &api.AppConfig{
		AppName: "demo",
		Group:   "dev",
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
	configService := app.NewConfigService("/tmp/nacos/config")
	//可以直接转化为底层http的api，对应nacos的openapi
	//namingService := app.NewNamingService()
	//namingHttpClient := namingService.HttpClient()
	configHttpClient := configService.HttpClient()

	r, er := configHttpClient.PublishConfig(&types.PublishConfig{
		DataID:  "demo.properties",
		Group:   "app1:beta",
		Tenant:  "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		Content: "text=hello,world",
	})
	if er != nil {
		panic(er)
	}
	fmt.Printf("result:%v\n", r.Success)
}
