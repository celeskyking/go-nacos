package service

import (
	"fmt"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/config"
	"github.com/celeskyking/go-nacos/pkg/pool"
	"github.com/celeskyking/go-nacos/pkg/util"
	"testing"
	"time"
)

func configS() ConfigService {
	op := &api.ConfigOption{
		Env:         "beta",
		AppName:     "app1",
		Namespace:   "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		Addresses:   []string{"127.0.0.1:8848"},
		LBStrategy:  api.RoundRobin,
		SnapshotDir: "/tmp/nacos/config/",
	}
	c := NewConfigService(op)
	return c
}

func TestNewConfigService(t *testing.T) {
	c := configS()
	mapFile, er := c.Properties("demo.properties")
	if er != nil {
		panic(er)
	}
	name := mapFile.MustGet("text")
	fmt.Printf("name:%s\n", name)
}

func TestConfigService_Properties(t *testing.T) {
	cs := configS()
	p, er := cs.Properties("demo.properties")
	if er != nil {
		panic(er)
	}
	v := p.MustGet("text")
	fmt.Println(v)
	pool.CloseAndWait()
}

func TestConfigService_Watch(t *testing.T) {
	c := configS()
	mapFile, er := c.Properties("demo.properties")
	if er != nil {
		panic(er)
	}
	mapFile.ListenValue("text", func(key string, curValue, newValue string, ctx *config.FileDesc) {
		fmt.Println("new value:" + newValue)
	})
	c.Watch()
	i := 0
	for i < 120 {
		time.Sleep(time.Second)
		i++
	}
	c.StopWatch()
}

func TestMD5(t *testing.T) {
	text := "text=hello,world6"
	fmt.Println(util.MD5([]byte(text)))
}
