package discovery

import (
	"fmt"
	"github.com/celeskyking/go-nacos/api"
	"testing"
)

func TestNewDiscoveryClient(t *testing.T) {
	config := &api.ConfigOption{
		AppName:    "demo",
		Env:        "dev",
		Namespace:  "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		Addresses:  []string{"127.0.0.1:8848"},
		LBStrategy: api.RoundRobin,
		Port:       8848,
	}
	c, er := NewDiscoveryClient(config)
	if er != nil {
		panic(er)
	}
	//time.Sleep(time.Hour)
	n, services, er := c.Naming().GetAllServices("7df0358d-8c73-4af3-8798-a54dd49aad7f")
	if er != nil {
		panic(er)
	}
	fmt.Printf("service num:%d", n)
	for _, s := range services {
		fmt.Printf("service :%+v", s)
	}
	er = c.Deregister()
	if er != nil {
		panic(er)
	}
}
