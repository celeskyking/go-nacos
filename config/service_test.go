package config

import (
	"fmt"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/pkg/pool"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"testing"
	"time"
)

func configS() ConfigService {
	op := &api.ConfigOptions{
		SnapshotDir: "/tmp/nacos/config/",
		ServerOptions: &api.ServerOptions{
			Addresses:  []string{"127.0.0.1:8848"},
			LBStrategy: api.RoundRobin,
		},
	}
	c := NewConfigService(op)
	return c
}

func TestNewConfigService(t *testing.T) {
	c := configS()
	mapFile, er := c.Properties(DefaultGroup, "demo.properties")
	if er != nil {
		panic(er)
	}
	name := mapFile.MustGet("text")
	fmt.Printf("name:%s\n", name)
}

func TestConfigService_Properties(t *testing.T) {
	cs := configS()
	p, er := cs.Properties(DefaultGroup, "demo.properties")
	if er != nil {
		panic(er)
	}
	v := p.MustGet("text")
	fmt.Println(v)
	pool.CloseAndWait()
}

func TestConfigService_Watch(t *testing.T) {
	c := configS()
	mapFile, er := c.Properties(DefaultGroup, "demo.properties")
	if er != nil {
		panic(er)
	}
	mapFile.ListenValue("text", func(key string, curValue, newValue string, ctx *types.FileDesc) {
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
