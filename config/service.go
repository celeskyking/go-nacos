package config

import (
	"context"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/api/cs"
	v1 "github.com/celeskyking/go-nacos/api/cs/v1"
	"github.com/celeskyking/go-nacos/config/converter"
	"github.com/celeskyking/go-nacos/config/converter/loader"
	"github.com/celeskyking/go-nacos/config/converter/properties"
	"github.com/celeskyking/go-nacos/err"
	"github.com/celeskyking/go-nacos/pkg/pool"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

const (
	DefaultGroup string = "DEFAULT_GROUP"
)

type ConfigService interface {
	//获取Properties文件
	Properties(group, file string) (*properties.MapFile, error)
	//文件
	Custom(group, file string, c converter.FileConverter) (cs.FileMirror, error)

	Watch()

	StopWatch()

	HttpClient() v1.ConfigHttpClient
}

func NewConfigService(options *api.ConfigOptions) ConfigService {
	httpOption := api.DefaultOption()
	httpOption.Servers = options.Addresses
	httpOption.LBStrategy = options.LBStrategy
	httpClient := v1.NewConfigHttpClient(httpOption)
	var loaders []loader.Loader
	localLoader := loader.NewLocalLoader(options.SnapshotDir)
	loaders = append(loaders, loader.NewRemoteLoader(httpClient))
	loaders = append(loaders, localLoader)
	return &configService{
		fileNotifier:   make(map[string]chan []byte, 0),
		fileVersion:    make(map[string]string, 0),
		status:         false,
		SnapshotDir:    options.SnapshotDir,
		loaders:        loaders,
		httpClient:     httpClient,
		NameSpaceID:    options.NamespaceID,
		snapshotWriter: localLoader.(loader.SnapshotWriter),
	}
}

type configService struct {
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

	watched bool

	NameSpaceID string
}

func (c *configService) Properties(group, file string) (*properties.MapFile, error) {
	f, er := c.Custom(group, file, converter.GetConverter("properties"))
	if er != nil {
		return nil, er
	}
	return f.(*properties.MapFile), nil
}

func (c *configService) HttpClient() v1.ConfigHttpClient {
	return c.httpClient
}

func (c *configService) getFile(group, file string) ([]byte, error) {
	desc := &types.FileDesc{
		Name:      file,
		Namespace: c.NameSpaceID,
		Group:     group,
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
			logrus.Errorf("file not found, file:%+v", desc)
			continue
		}
	}
	return nil, err.ErrFileNotFound
}

func (c *configService) Custom(group, file string, converter converter.FileConverter) (cs.FileMirror, error) {
	g := group
	if g == "" {
		g = DefaultGroup
	}
	bs, er := c.getFile(g, file)
	if er != nil {
		return nil, er
	}
	f := converter.Convert(&types.FileDesc{
		Namespace: c.NameSpaceID,
		Group:     g,
		Name:      file,
	}, bs)
	m := util.MD5(bs)
	k := buildFileKey(c.NameSpaceID, g, file)
	if _, ok := c.fileNotifier[file]; !ok {
		//100长度的缓冲队列
		c.fileNotifier[k] = make(chan []byte, 100)

	}
	go f.OnChanged(c.fileNotifier[file])
	c.fileVersion[k] = m
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.watched {
		c.Watch()
		c.watched = true
	}
	return f, nil
}

func (c *configService) Watch() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.watched {
		return
	}
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
			if len(list) == 0 {
				time.Sleep(5 * time.Second)
			}
			changes, er := c.httpClient.ListenConfigs(&types.ListenConfigsRequest{
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
					desc := &types.FileDesc{
						Namespace: c.NameSpaceID,
						Name:      k.DataID,
						Group:     k.Group,
					}
					pool.Go(func(context context.Context) {
						c.flushSnapshot(desc, vb)
					})
					if notifyC, ok := c.fileNotifier[k.Line()]; ok {
						tmp := make([]byte, len(vb))
						copy(tmp, vb)
						m := util.MD5(tmp)
						c.fileVersion[k.Line()] = m
						notifyC <- vb
					}
				}
			}
			reties = 0
		}
	}()
}

func (c *configService) flushSnapshot(desc *types.FileDesc, content []byte) {
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

func buildFileKey(namespace, group, file string) string {
	if namespace == "" {
		namespace = "Public"
	}
	key := &types.ListenKey{
		Tenant: namespace,
		Group:  group,
		DataID: file,
	}
	return key.Line()
}

func (c *configService) listenKeys() ([]*types.ListenKey, error) {
	var keys []*types.ListenKey
	if len(c.fileNotifier) == 0 {
		return nil, nil
	}
	for k := range c.fileNotifier {
		listenKey, er := types.ParseListenKey(k)
		if er != nil {
			return nil, er
		}
		listenKey.ContentMD5 = c.fileVersion[k]
		keys = append(keys, listenKey)
	}
	return keys, nil
}
