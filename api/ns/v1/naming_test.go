package v1

import (
	"fmt"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"testing"
)

var Naming NamingHttpClient

func init() {
	op := api.DefaultOption()
	op.Servers = []string{"http://127.0.0.1:8848"}
	op.LBStrategy = api.RoundRobin
	Naming = NewNamingHttpClient(op)
}

func TestNamingHttpClient_CatalogServices(t *testing.T) {
	result, er := Naming.CatalogServices(true, "7df0358d-8c73-4af3-8798-a54dd49aad7f")
	if er != nil {
		t.Errorf("catalog services failed:%+v", er)
		t.Fail()
	}
	for _, s := range result {
		fmt.Printf("serviceDetail:%s", util.ToJSONString(s))
	}
}

func TestNamingHttpClient_CreateService(t *testing.T) {
	result, err := Naming.CreateService(&types.Service{
		NamespaceId:      "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		ServiceName:      "go-nacos",
		GroupName:        "dev",
		ProtectThreshold: 0.6,
		Metadata: util.MapToString(map[string]string{
			"name": "tianqing.wang",
			"age":  "30",
		}),
	})
	if err != nil {
		t.Error("create service failed", err)
		t.Fail()
	}
	fmt.Printf("create service:%v\n", result.Success)
}

func TestNamingHttpClient_DeleteService(t *testing.T) {
	result, err := Naming.DeleteService(&types.Service{
		NamespaceId: "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		ServiceName: "go-nacos",
		GroupName:   "dev",
	})
	if err != nil {
		t.Error("delete service failed", err)
		t.Fail()
	}
	fmt.Printf("delete service:%v\n", result.Success)
}

func TestNamingHttpClient_GetLeader(t *testing.T) {
	result, err := Naming.GetLeader()
	if err != nil {
		t.Error("get leader failed", err)
		t.Fail()
	}
	fmt.Printf("get leader:%+v\n", result)
}

func TestNamingHttpClient_GetMetrics(t *testing.T) {
	result, err := Naming.GetMetrics()
	if err != nil {
		t.Error("get metrics failed", err)
		t.Fail()
	}
	fmt.Printf("get metrics:%+v\n", result)
}

func TestNamingHttpClient_GetNacosServers(t *testing.T) {
	result, err := Naming.GetNacosServers()
	if err != nil {
		t.Error("get servers failed", err)
		t.Fail()
	}
	fmt.Printf("get servers :%s", util.ToJSONString(result))
}

func TestNamingHttpClient_GetService(t *testing.T) {
	result, err := Naming.GetService(&types.Service{
		NamespaceId: "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		ServiceName: "demo",
		GroupName:   "dev",
	})
	if err != nil {
		t.Error("create service failed", err)
		t.Fail()
	}
	fmt.Printf("create service: %s", util.ToJSONString(result))
}

func TestNamingHttpClient_GetSwitches(t *testing.T) {
	result, err := Naming.GetSwitches()
	if err != nil {
		t.Error("get switches failed", err)
		t.Fail()
	}
	fmt.Printf("get switches :%s", util.ToJSONString(result))
}

func TestNamingHttpClient_RegisterServiceInstance(t *testing.T) {
	result, err := Naming.RegisterServiceInstance(&types.ServiceInstance{
		GroupName:   "beta",
		ServiceName: "local-2",
		IP:          "10.10.10.15",
		Port:        8080,
		Metadata: util.MapToString(map[string]string{
			"name": "tianqing.wang",
			"age":  "30",
		}),
		Weight:    1.0,
		Healthy:   true,
		Enable:    true,
		Ephemeral: false,
	})
	if err != nil {
		t.Error("create service failed", err)
		t.Fail()
	}
	fmt.Printf("create service: %s", util.ToJSONString(result))
}

func TestNamingHttpClient_PatchCluster(t *testing.T) {
	cluster := &types.Cluster{
		NamespaceID:           "",
		ServiceName:           "local-2",
		ClusterName:           "DEFAULT",
		GroupName:             "beta",
		UseInstancePort4Check: true,
		CheckPort:             8080,
		HealthChecker:         types.NewNoneHealthChecker(),
	}
	r, er := Naming.PatchCluster(cluster)
	if er != nil {
		t.Error("patch cluster failed", er)
		t.Fail()
	}
	fmt.Printf("patch cluster:%v", r.Success)
}

func TestNamingHttpClient_UpdateServiceInstanceHealthy(t *testing.T) {
	r, er := Naming.UpdateServiceInstanceHealthy(&types.UpdateServiceInstanceHealthyRequest{
		NamespaceID: "",
		ServiceName: "local-2",
		ClusterName: "DEFAULT",
		GroupName:   "beta",
		IP:          util.LocalIP(),
		Port:        8080,
		Healthy:     true,
	})
	if er != nil {
		panic(er)
	}
	fmt.Printf("update healthy:%v", r.Success)
}
