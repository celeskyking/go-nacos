package v1

import (
	"fmt"
	"gitlab.mfwdev.com/portal/go-nacos/api"
	"gitlab.mfwdev.com/portal/go-nacos/types"
	"testing"
	"time"
)

var ConfigClient ConfigHttpClient

func init() {
	op := api.DefaultOption()
	op.Servers = []string{"http://127.0.0.1:8848"}
	op.LBStrategy = api.RoundRobin
	ConfigClient = NewConfigHttpClient(op)
}

func TestConfigHttpClient_PublishConfig(t *testing.T) {
	pr, er := ConfigClient.PublishConfig(&types.PublishConfig{
		DataID:  "demo.properties",
		Group:   "app1:beta",
		Tenant:  "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		Content: "text=hello,world",
	})
	if er != nil {
		panic(er)
	}
	fmt.Println(pr.Success)
}

func TestConfigHttpClient_GetConfigs(t *testing.T) {
	r, er := ConfigClient.GetConfigs(&types.ConfigsRequest{
		DataID: "demo.properties",
		Group:  "app1:beta",
		Tenant: "7df0358d-8c73-4af3-8798-a54dd49aad7f",
	})
	if er != nil {
		panic(er)
	}
	fmt.Println(r.Value)
}

func TestConfigHttpClient_DeleteConfigs(t *testing.T) {
	r, er := ConfigClient.DeleteConfigs(&types.ConfigsRequest{
		DataID: "demo.properties",
		Group:  "app1:beta",
		Tenant: "7df0358d-8c73-4af3-8798-a54dd49aad7f",
	})
	if er != nil {
		panic(er)
	}
	fmt.Println(r.Success)
}

func TestConfigHttpClient_ListenConfigs(t *testing.T) {
	key := &types.ListenKey{
		DataID: "jdbc.properties",
		Group:  "DEFAULT_GROUP",
	}
	l := &types.ListenConfigsRequest{ListeningConfigs: []*types.ListenKey{key}}
	fmt.Println(l.Line())
	changes, er := ConfigClient.ListenConfigs(l)
	if er != nil {
		panic(er)
	}
	for _, c := range changes {
		fmt.Println("测试监听:" + c.NewValue)
	}
	i := 0
	for i < 30 {
		time.Sleep(time.Second)
		i++
	}
}
