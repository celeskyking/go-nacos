package discovery

import (
	"fmt"
	"gitlab.mfwdev.com/portal/go-nacos/api"
	"gitlab.mfwdev.com/portal/go-nacos/naming"
	"testing"
	"time"
)

var client *Client

func init() {
	config := &api.ServerOptions{
		Addresses:       []string{"127.0.0.1:8848"},
		LBStrategy:      api.RoundRobin,
		EndpointEnabled: false,
	}

	appConfig := &api.DiscoveryOptions{
		AppName:   "demo",
		Env:       "dev",
		Namespace: "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		Cluster:   "",
		Port:      8848,
	}
	ns := naming.NewNamingService(config)
	client = NewDiscoveryClient(ns, appConfig)
	er := client.Register()
	if er != nil {
		panic(er)
	}
}

func TestNewDiscoveryClient(t *testing.T) {
	c := client
	er := c.Register()
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

func TestClient_GetInstances(t *testing.T) {
	sl, er := client.GetInstances("local-2", &naming.QueryOptions{
		Group: "beta",
		Watch: true,
	})
	if er != nil {
		panic(er)
	}
	for _, s := range sl.GetAll() {
		fmt.Printf("instance : %+v\n", s)
	}
	for {
		time.Sleep(time.Second)
	}
}
