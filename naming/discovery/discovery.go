package discovery

import (
	"github.com/celeskyking/go-nacos/api"
	beat "github.com/celeskyking/go-nacos/naming/heartbeat"
	"github.com/celeskyking/go-nacos/naming/service"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"github.com/pkg/errors"
)

type Client struct {
	IP string
	//当前服务的端口信息
	Port int

	Namespace string

	AppName string

	Env string

	heartBeatService beat.HeartBeatService

	naming service.NamingService

	instance *types.ServiceInstance

	stopC chan struct{}
}

func NewDiscoveryClient(config *api.ConfigOption) (*Client, error) {
	ip := util.LocalIP()
	naming := service.NewNamingService(config)
	c := &Client{
		IP:        ip,
		Port:      config.Port,
		Namespace: config.Namespace,
		AppName:   config.AppName,
		Env:       config.Env,
		stopC:     make(chan struct{}, 0),
		naming:    naming,
	}
	er := c.Register()
	if er != nil {
		return nil, er
	}
	return c, nil
}

func (c *Client) Naming() service.NamingService {
	return c.naming
}

func (c *Client) Register() error {
	req := &types.ServiceInstance{
		IP:          c.IP,
		Port:        c.Port,
		NamespaceID: c.Namespace,
		ServiceName: c.AppName,
		GroupName:   c.Env,
		Ephemeral:   true,
		Weight:      1.0,
		Healthy:     true,
		Enable:      true,
	}
	er := c.naming.RegisterInstance(req)
	if er != nil {
		return errors.Wrap(er, " register service")
	}
	h := beat.NewHeartBeatService(c.naming.HttpClient(), req)
	c.instance = req
	go h.Start(c.stopC)
	c.heartBeatService = h
	return nil
}

func (c *Client) Deregister() error {
	c.stopC <- struct{}{}
	er := c.naming.DeRegisterInstance(c.instance)
	if er != nil {
		return errors.Wrap(er, "deregister service")
	}
	c.naming.Stop()
	return nil
}
