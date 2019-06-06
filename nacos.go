package nacos

import (
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/config"
	"github.com/celeskyking/go-nacos/naming"
	"github.com/celeskyking/go-nacos/naming/discovery"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var ErrSnapshotDirIsEmpty = errors.New("未指定config service的snapshot目录")

//NacosFactory nacos工厂
type Factory struct {
}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) NewConfigService(options *api.ConfigOptions) config.ConfigService {
	if options.SnapshotDir == "" {
		logrus.Errorf("未指定config service的snapshot的目录")
		panic(ErrSnapshotDirIsEmpty)
	}
	return config.NewConfigService(options)
}

func (f *Factory) NewNamingService(options *api.ServerOptions) naming.NamingService {
	return naming.NewNamingService(options)
}

type Application struct {
	Config *api.AppConfig

	configServers *api.ServerOptions

	namingServers *api.ServerOptions
}

func NewApplication(appConfig *api.AppConfig) *Application {
	return &Application{
		Config: appConfig,
	}
}

func (a *Application) SetConfigServers(options *api.ServerOptions) {
	a.configServers = options
}

func (a *Application) SetNamingServers(options *api.ServerOptions) {
	a.namingServers = options
}

func (a *Application) SetServers(options *api.ServerOptions) {
	a.namingServers = options
	a.configServers = options
}

func (a *Application) NewConfigService(snapshotDir string) config.ConfigService {
	if a.configServers == nil {
		panic(errors.New("未配置nacos server信息"))
	}
	return config.NewConfigService(&api.ConfigOptions{
		ServerOptions: a.configServers,
		SnapshotDir:   snapshotDir,
		AppName:       a.Config.AppName,
		Env:           a.Config.Env,
		Cluster:       a.Config.Cluster,
		Namespace:     a.Config.Namespace,
	})
}

func (a *Application) NewNamingService() naming.NamingService {
	if a.namingServers == nil {
		panic(errors.New("未配置nacos server信息"))
	}
	return naming.NewNamingService(a.namingServers)
}

func (a *Application) NewDiscoveryClient() *discovery.Client {
	if a.Config.IP == "" {
		a.Config.IP = util.LocalIP()
	}
	return discovery.NewDiscoveryClient(a.NewNamingService(), &api.DiscoveryOptions{
		IP:        a.Config.IP,
		Namespace: a.Config.Namespace,
		AppName:   a.Config.AppName,
		Cluster:   a.Config.Cluster,
		Env:       a.Config.Env,
		Port:      a.Config.Port,
	})
}
