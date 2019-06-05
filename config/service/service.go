package service

import (
	"context"
	"github.com/celeskyking/go-nacos/api"
	v1 "github.com/celeskyking/go-nacos/api/cs/v1"
	"github.com/celeskyking/go-nacos/config"
	"github.com/celeskyking/go-nacos/config/converter/loader"
	"github.com/celeskyking/go-nacos/config/converter/properties"
	"github.com/celeskyking/go-nacos/err"
	"github.com/celeskyking/go-nacos/pkg/pool"
	"github.com/celeskyking/go-nacos/pkg/util"
	types2 "github.com/celeskyking/go-nacos/types"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
	"time"
)

type ConfigService interface {

	//获取Properties文件
	Properties(file string) (*properties.MapFile, error)

	//文件
	Custom(file string, c config.FileConverter) (config.FileMirror, error)

	Watch()

	StopWatch()
}

func NewConfigService(option *api.ConfigOption) ConfigService {
	httpOption := api.DefaultOption()
	httpOption.Servers = option.Addresses
	httpOption.LBStrategy = option.LBStrategy
	httpClient := v1.NewConfigHttpClient(httpOption)
	var loaders []loader.Loader
	localLoader := loader.NewLocalLoader(option.SnapshotDir)
	loaders = append(loaders, loader.NewRemoteLoader(httpClient))
	loaders = append(loaders, localLoader)
	return &configService{
		Env:            option.Env,
		Namespace:      option.Namespace,
		AppName:        option.AppName,
		fileNotifier:   make(map[string]chan []byte, 0),
		fileVersion:    make(map[string]string, 0),
		status:         false,
		SnapshotDir:    option.SnapshotDir,
		loaders:        loaders,
		httpClient:     httpClient,
		snapshotWriter: localLoader.(loader.SnapshotWriter),
	}
}

type configService struct {
	//数据中心
	Env string
	//对应的命名空间
	Namespace string
	//对应的group
	AppName string
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
	//loader
	loaders []loader.Loader

	SnapshotDir string

	httpClient v1.ConfigHttpClient

	snapshotWriter loader.SnapshotWriter
}

func (c *configService) Properties(file string) (*properties.MapFile, error) {
	f, er := c.Custom(file, config.GetConverter("properties"))
	if er != nil {
		return nil, er
	}
	return f.(*properties.MapFile), nil
}

func (c *configService) getFile(file string) ([]byte, error) {
	desc := &config.FileDesc{
		Name:      file,
		Namespace: c.Namespace,
		AppName:   c.AppName,
		Env:       c.Env,
	}
	for _, l := range c.loaders {
		data, er := l.Load(desc)
		if er == nil {
			switch l.(type) {
			case *loader.LocalLoader:
				logrus.Infof("loader: localLoader")
			case *loader.RemoteLoader:
				logrus.Info("loader: remoteLoader")
				pool.Go(func(ctx context.Context) {
					er = c.snapshotWriter.Write(desc, data)
					if er != nil {
						logrus.Errorf("flush snapshot error:%+v", er)
					}
				})
			}
			return data, nil
		} else {
			logrus.Errorf("load error:%+v", er)
			continue
		}
	}
	return nil, err.ErrLoaderNotWork
}

func (c *configService) Custom(file string, converter config.FileConverter) (config.FileMirror, error) {
	bs, er := c.getFile(file)
	if er != nil {
		return nil, er
	}
	f := converter.Convert(&config.FileDesc{
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
			changes, er := c.httpClient.ListenConfigs(&types2.ListenConfigsRequest{
				ListeningConfigs: list,
			})
			if er != nil {
				logrus.Errorf("listen to nacos error:%+v", er)
				reties = reties + 1
				time.Sleep(time.Duration(util.Min(reties*5, maxDelay)) * time.Second)
				return
			}
			for _, change := range changes {
				k := change.Key
				v := change.NewValue
				if v != "" {
					k.ContentMD5 = ""
					vb := []byte(v)
					parts := strings.Split(k.Group, ":")
					desc := &config.FileDesc{
						Namespace: c.Namespace,
						Name:      k.DataID,
						AppName:   parts[0],
						Env:       parts[1],
					}
					pool.Go(func(context context.Context) {
						c.flushSnapshot(desc, vb)
					})
					if notifyC, ok := c.fileNotifier[k.Line()]; ok {
						tmp := make([]byte, len(vb))
						copy(tmp, vb)
						c.fileVersion[k.Line()] = util.MD5(tmp)
						notifyC <- vb
					}
				}
			}
			reties = 0
		}
	}()
}

func (c *configService) flushSnapshot(desc *config.FileDesc, content []byte) {
	er := c.snapshotWriter.Write(desc, content)
	if er != nil {
		logrus.Errorf("snapshot flush content error:%+v", er)
	}
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
