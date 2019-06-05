# Go-Nacos

当前还处于开发阶段,只完成了config部分的代码



## 快速开始


```go
    op := &ConfigOption{
		Env:        "beta",
		AppName:    "app1",
		Namespace:  "7df0358d-8c73-4af3-8798-a54dd49aad7f",
		Addresses:  []string{"127.0.0.1:8848"},
		LBStrategy: cs.RoundRobin,
	}
    //创建ConfigService 
    c := service.NewConfigService(op)
	//新建properties文件,目前只支持properties格式的文件
	mapFile, er  := c.Properties("demo.properties")
    if er != nil {
        panic(er)
    }
	//获取value
    name := mapFile.MustGet("text")
    fmt.Printf("name:%s\n",name)
    //监听值变化
    mapFile.ListenValue("text", func(key string, curValue, newValue string, ctx *types.FileDesc) {
        fmt.Println("new value:"+newValue)
    })
    c.Watch()
```