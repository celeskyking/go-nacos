# Go-Nacos

当前还处于开发阶段,只完成了config部分的代码



## 快速开始


```go
import (
	"fmt"
	"github.com/celeskyking/go-nacos"
	"github.com/celeskyking/go-nacos/api"
)

func main() {
	config := &api.ConfigOption{
		AppName:    "demo",
		Env:        "dev",
		Namespace:  "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		Addresses:  []string{"127.0.0.1:8848"},
		LBStrategy: api.RoundRobin,
		Port:       8080,
	}
	f := nacos.NewNacosFactory(config)
	dc := f.NewDiscoveryClient()
	dc.SetEphemeral(true)
	er := dc.Register()
	if er != nil {
		panic(er)
	}
	n, sl, er := dc.Naming().GetAllServices("7df0358d-8c73-4af3-8798-a54dd49aad7f")
	if er != nil {
		panic(er)
	}
	fmt.Printf("service num:%d", n)
	for _, s := range sl {
		fmt.Println(s)
	}
	er = dc.Deregister()
	if er != nil {
		panic(er)
	}
}
```