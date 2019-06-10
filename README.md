# Go-Nacos


## 简介
Nacos的Go客户端,不仅封装了OpenApi，参考Nacos的Java Client封装了一些高级功能，加入了Endpoint的支持，支持针对于Nacos Servers
的客户端负载均衡能力。ConfigServer支持本地快照。NamingService支持AP和CP模式,支持自定义HealthCheck等。


## 快速开始


### Config


ConfigHttpClient 实现了Nacos 1.0.0的OpenApi的接口能力,
ConfigService是一个对于OpenApi的高级封装,定义了一些最佳实践的api

#### 概念

* AppName  
    整体思路是以应用为中心,对齐Discovery和ConfigService的概念，AppName:Env对应Nacos的Group概念
* Env
    当前的应用的所归属的环境信息,与AppName配合使用,对应Nacos的Group的概念
* Namespace
    租户的概念，概念对应Nacos的Namespace


#### 例子

```go
import (
	"fmt"
	"github.com/celeskyking/go-nacos"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/config"
	"github.com/celeskyking/go-nacos/pkg/util"
)

//
//name=go-nacos
//size=20
//
//
//
func main() {
	appConfig := &api.AppConfig{
		AppName:"demo",
		Env:"dev",
		Port:8080,
		//IP可以为空
		IP:util.LocalIP(),
	}
	app := nacos.NewApplication(appConfig)
	configService := app.NewConfigService("/tmp/nacos/config/snapshot")
	//目前只支持properties文件,不过支持自定义格式文件的扩展,Custom方法
	file, er := configService.Properties("demo.properties")
	if er != nil {
		panic(er)
	}
	file.ListenValue("name", func(key string, curValue, newValue string, ctx *config.FileDesc) {
		fmt.Println("new value:"+newValue)
	})
	n := file.MustGet("name")
	fmt.Printf("name:%s",n)
	s := file.MustGetInt32("size")
	fmt.Printf("size:%d\n",s)
	configService.StopWatch()
}
```



### NamingService

NamingHttpClient 实现了Nacos 1.0.0的OpenApi的接口能力,
NamingService是一个简单封装,把服务发现和注册的能力统一封装了一下。


#### 概念

* AppName  
    整体思路是以应用为中心，AppName:Env对应Nacos的Group概念
* Env
    当前的应用的所归属的环境信息,与AppName配合使用,对应Nacos的Group的概念
* Namespace
    租户的概念，概念对应Nacos的Namespace
* Cluster
    实例的分组关键字,概念对齐Nacos的Cluster

   

```go

import (
	"github.com/celeskyking/go-nacos"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
)

func main() {
	appConfig := &api.AppConfig{
		AppName:"demo",
		Env:"dev",
		Port:8080,
		//IP可以为空
		IP:util.LocalIP(),
	}
	app := nacos.NewApplication(appConfig)
	ns := app.NewNamingService()
	er := ns.RegisterInstance(&types.ServiceInstance{
		GroupName:   "beta",
		ServiceName: "local-2",
		IP:          "10.10.10.15",
		Port:        8080,
		Metadata: util.MapToString(map[string]string{
			"name": "go-nacos",
			"age":  "30",
		}),
		Weight:    1.0,
		Healthy:   true,
		Enable:    true,
		Ephemeral: false,
	})
	if er != nil {
		panic(er)
	}
}


```


### Discovery


```go
import (
	"fmt"
	"github.com/celeskyking/go-nacos"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/naming"
	"github.com/celeskyking/go-nacos/pkg/util"
)

func main() {
	appConfig := &api.AppConfig{
		AppName:"demo",
		Env:"dev",
		Port:8080,
		//IP可以为空
		IP:util.LocalIP(),
	}
	app := nacos.NewApplication(appConfig)
	dc := app.NewDiscoveryClient()
	er := dc.Register()
	if er != nil {
		panic(er)
	}
	serverList, er := dc.GetInstances("app2", &naming.QueryOptions{
		Namespace:"",
		Cluster:"",
		//会接受推送
		Watch:true,
	})
	if er != nil {
		panic(er)
	}
	instance  := serverList.SelectOne()
	fmt.Println(util.ToJSONString(instance))
	all := serverList.GetAll()
	fmt.Println(util.ToJSONString(all))
	serverList.StopListen()
	er = dc.Deregister()
	if er != nil {
		panic(er)
	}
}

```


### http


```go

import (
	"fmt"
	"github.com/celeskyking/go-nacos"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
)

func main () {
	appConfig := &api.AppConfig{
		AppName:"demo",
		Env:"dev",
		Port:8080,
		//IP可以为空
		IP:util.LocalIP(),
	}
	app := nacos.NewApplication(appConfig)
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
	fmt.Printf("result:%v\n",r.Success)
}
```


### 功能列表

#### Config

* 支持Config本地快照
* 支持Nacos Server端的健康监测
* 支持Endpoint
#### NamingService
* 支持高级Api(discovery)
* 支持全部OpenApi
* 支持服务列表的Push
* 支持Endpoint
* 支持Nacos Server端的健康监测
