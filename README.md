# Go-Nacos

当前还处于开发阶段,只完成了config部分的代码



## 快速开始


### Config


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
