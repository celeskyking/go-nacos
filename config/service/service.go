package service

import (
	config2 "github.com/celeskyking/go-nacos/api/cs"
	v1 "github.com/celeskyking/go-nacos/api/cs/v1"
	"github.com/celeskyking/go-nacos/config"
	"github.com/celeskyking/go-nacos/config/converter/properties"
	"github.com/celeskyking/go-nacos/config/types"
	"github.com/celeskyking/go-nacos/pkg/util"
	types2 "github.com/celeskyking/go-nacos/types"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

type ConfigService interface {

	//获取Properties文件
	Properties(file string) (*properties.MapFile, error)

	//文件
	Custom(file string, c config.FileConverter) (types.FileMirror, error)

	Watch()

	StopWatch()
}

func NewConfigService(option *ConfigOption) ConfigService {
	httpOption := v1.DefaultOption()
	httpOption.Servers = option.Addresses
	httpOption.LBStrategy = option.LBStrategy
	return &configService{
		Env:          option.Env,
		Namespace:    option.Namespace,
		AppName:      option.AppName,
		httpClient:   v1.NewConfigHttpClient(httpOption),
		fileNotifier: make(map[string]chan []byte, 0),
		fileVersion:  make(map[string]string, 0),
		status:       false,
	}
}

type configService struct {
	//数据中心
	Env string
	//对应的命名空间
	Namespace string
	//对应的group
	AppName string
	//http请求的端口
	httpClient v1.ConfigHttpClient
	//文件监听器,key为ListenKey
	fileNotifier map[string]chan []byte
	//文件的版本
	fileVersion map[string]string
	//锁
	lock sync.Mutex
	//true为开启，false为关闭
	status bool
	//开始
	listening chan struct{}
}

func (c *configService) Properties(file string) (*properties.MapFile, error) {
	f, er := c.Custom(file, config.GetConverter("properties"))
	if er != nil {
		return nil, er
	}
	return f.(*properties.MapFile), nil
}

func (c *configService) Custom(file string, converter config.FileConverter) (types.FileMirror, error) {
	bs, er := c.getFile(file)
	if er != nil {
		return nil, er
	}
	f := converter.Convert(&types.FileDesc{
		Namespace: c.Namespace,
		AppName:   c.AppName,
		Env:       c.Env,
		Name:      file,
	}, bs)
	m := util.MD5(bs)
	k := buildFileKey(c.Namespace, c.AppName, c.Env, file)
	if _, ok := c.fileNotifier[file]; !ok {
		//100长度的缓冲队列
		c.fileNotifier[k] = make(chan []byte, 100)

	}
	go f.OnChanged(c.fileNotifier[file])
	c.fileVersion[k] = m
	return f, nil
}

func (c *configService) getFile(file string) (b []byte, err error) {
	var bs []byte
	c.httpClient.GetConfigs(&types2.ConfigsRequest{
		DataID: file,
		Tenant: c.Namespace,
		Group:  c.group(),
	}, func(response *types2.ConfigsResponse, er error) {
		if er != nil {
			err = er
			return
		}
		bs = []byte(response.Value)
		b = bs
	})
	return
}

func (c *configService) group() string {
	return c.AppName + ":" + c.Env
}

func (c *configService) Watch() {
	go func() {
		reties := 0
		maxDelay := 60
		c.status = true
		for c.status {
			list, er := c.listenKeys()
			if er != nil {
				logrus.Errorf("listen nacos file error:%+v", er)
				os.Exit(1)
			}
			c.httpClient.ListenConfigs(&types2.ListenConfigsRequest{
				ListeningConfigs: list,
			}, func(result []*types2.ListenChange, err error) {
				if err != nil {
					logrus.Errorf("listen to nacos error:%+v", err)
					reties = reties + 1
					time.Sleep(time.Duration(util.Min(reties*5, maxDelay)) * time.Second)
					return
				}
				for _, change := range result {
					k := change.Key
					v := change.NewValue
					if v != "" {
						k.ContentMD5 = ""
						if notifyC, ok := c.fileNotifier[k.Line()]; ok {
							vb := []byte(v)
							tmp := make([]byte, len(vb))
							copy(tmp, vb)
							c.fileVersion[k.Line()] = util.MD5(tmp)
							notifyC <- vb
						}
					}
				}
				reties = 0
			})
		}
	}()
}

func (c *configService) StopWatch() {
	c.status = false
	for _, f := range c.fileNotifier {
		close(f)
	}
}

func buildFileKey(namespace, app, env, file string) string {
	key := &types2.ListenKey{
		Tenant: namespace,
		Group:  app + ":" + env,
		DataID: file,
	}
	return key.Line()
}

func (c *configService) listenKeys() ([]*types2.ListenKey, error) {
	var keys []*types2.ListenKey
	for k := range c.fileNotifier {
		listenKey, er := types2.ParseListenKey(k)
		if er != nil {
			return nil, er
		}
		listenKey.ContentMD5 = c.fileVersion[k]
		keys = append(keys, listenKey)
	}
	return keys, nil
}

type ConfigOption struct {
	//应用名称
	AppName string
	//环境名称
	Env string
	// 当前的namespace
	Namespace string
	//nacos 配置中心的地址
	Addresses []string
	//负载均衡策略
	LBStrategy config2.LBStrategy
}
