package main

import (
	"github.com/celeskyking/go-nacos"
	"github.com/celeskyking/go-nacos/api"
)

func main() {
	factory := nacos.NewFactory()
	factory.NewConfigService(&api.ConfigOptions{})
	factory.NewNamingService(&api.ServerOptions{})
}
