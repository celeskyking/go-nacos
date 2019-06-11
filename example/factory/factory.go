package main

import (
	"gitlab.mfwdev.com/portal/go-nacos"
	"gitlab.mfwdev.com/portal/go-nacos/api"
)

func main() {
	factory := nacos.NewFactory()
	factory.NewConfigService(&api.ConfigOptions{})
	factory.NewNamingService(&api.ServerOptions{})
}
