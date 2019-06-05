package nacos

import (
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/config"
	"github.com/celeskyking/go-nacos/naming"
	"github.com/celeskyking/go-nacos/naming/discovery"
	"sync"
)

//NacosFactory nacos工厂
type Factory struct {
	once sync.Once

	namingService naming.NamingService

	configService config.ConfigService
	//配置信息
	Config *api.ConfigOption

	lock sync.Mutex
}

func NewNacosFactory(config *api.ConfigOption) *Factory {
	return &Factory{
		Config: config,
	}
}

func (n *Factory) NamingService() naming.NamingService {
	if n.namingService == nil {
		n.lock.Lock()
		defer n.lock.Unlock()
		if n.namingService == nil {
			n.namingService = naming.NewNamingService(n.Config)
		}
	}
	return n.namingService
}

func (n *Factory) ConfigService() config.ConfigService {
	if n.configService == nil {
		n.lock.Lock()
		defer n.lock.Unlock()
		if n.configService == nil {
			n.configService = config.NewConfigService(n.Config)
		}
	}
	return n.configService
}

func (n *Factory) NewDiscoveryClient() *discovery.Client {
	ns := n.NamingService()
	return discovery.NewDiscoveryClient(ns)
}
