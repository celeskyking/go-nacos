package v1

import (
	"fmt"
	"github.com/celeskyking/go-nacos/api/cs"
	"github.com/celeskyking/go-nacos/types"
	"testing"
	"time"
)

var ConfigClient ConfigHttpClient

func init() {
	op := DefaultOption()
	op.Servers = []string{"http://127.0.0.1:8848"}
	op.LBStrategy = cs.RoundRobin
	ConfigClient = NewConfigHttpClient(op)
}

func TestConfigHttpClient_PublishConfig(t *testing.T) {
	ConfigClient.PublishConfig(&types.PublishConfig{
		DataID:  "demo.properties",
		Group:   "app1:beta",
		Tenant:  "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		Content: "text=hello,world",
	}, func(result *types.PublishResult, err error) {
		if err != nil {
			panic(err)
		}
		fmt.Println(result.Success)
	})
}

func TestConfigHttpClient_GetConfigs(t *testing.T) {
	ConfigClient.GetConfigs(&types.ConfigsRequest{
		DataID: "demo.properties",
		Group:  "app1:beta",
		Tenant: "7df0358d-8c73-4af3-8798-a54dd49aad7f",
	}, func(response *types.ConfigsResponse, err error) {
		if err != nil {
			panic(err)
		}
		fmt.Println(response.Value)
	})
}

func TestConfigHttpClient_DeleteConfigs(t *testing.T) {
	ConfigClient.DeleteConfigs(&types.ConfigsRequest{
		DataID: "demo.properties",
		Group:  "app1:beta",
		Tenant: "7df0358d-8c73-4af3-8798-a54dd49aad7f",
	}, func(response *types.DeleteResult, err error) {
		if err != nil {
			panic(err)
		}
		fmt.Println(response.Success)
	})
}

func TestConfigHttpClient_ListenConfigs(t *testing.T) {
	key := &types.ListenKey{
		DataID: "jdbc.properties",
		Group:  "DEFAULT_GROUP",
	}
	l := &types.ListenConfigsRequest{ListeningConfigs: []*types.ListenKey{key}}
	fmt.Println(l.Line())
	ConfigClient.ListenConfigs(l, func(result []*types.ListenChange, err error) {
		if err != nil {
			panic(err)
		}
		for _, c := range result {
			fmt.Println("测试监听:" + c.NewValue)
		}
	})
	i := 0
	for i < 30 {
		time.Sleep(time.Second)
		i++
	}
}
