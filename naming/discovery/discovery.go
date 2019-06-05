package discovery

import (
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/naming"
	beat "github.com/celeskyking/go-nacos/naming/heartbeat"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"github.com/pkg/errors"
)

type Client struct {
	IP string
	//当前服务的端口信息
	Port int

	Namespace string
	//Cluster集群
	Cluster string
	//应用名
	AppName string
	//环境名
	Env string
	//心跳服务
	heartBeatService beat.HeartBeatService

	naming naming.NamingService

	Ephemeral bool

	instance *types.ServiceInstance
	//初始化的健康状态
	InitHealthy bool

	stopC chan struct{}
}

func NewDiscoveryClient(service naming.NamingService) *Client {
	ip := util.LocalIP()
	config := service.GetConfig()
	c := &Client{
		IP:        ip,
		Port:      config.Port,
		Namespace: config.Namespace,
		Cluster:   config.Cluster,
		AppName:   config.AppName,
		Env:       config.Env,
		stopC:     make(chan struct{}, 0),
		naming:    service,
	}
	return c
}

func NewDiscoveryClientFromConfig(config *api.ConfigOption) *Client {
	naming := naming.NewNamingService(config)
	return NewDiscoveryClient(naming)
}

func (c *Client) Naming() naming.NamingService {
	return c.naming
}

func (c *Client) SetInitHealthy(healthy bool) {
	c.InitHealthy = healthy
}

func (c *Client) SetEphemeral(ephemeral bool) {
	c.Ephemeral = ephemeral
}

func (c *Client) SetHealthy(healthy bool) error {
	return c.naming.SetInstanceHealthy(c.AppName, &naming.QueryOptions{
		Namespace: c.Namespace,
		Group:     c.Env,
		Cluster:   c.Cluster,
		Healthy:   healthy,
	})
}

func (c *Client) Register() error {
	req := &types.ServiceInstance{
		IP:          c.IP,
		Port:        c.Port,
		NamespaceID: c.Namespace,
		ServiceName: c.AppName,
		GroupName:   c.Env,
		ClusterName: c.Cluster,
		Ephemeral:   true,
		Weight:      1.0,
		Healthy:     true,
		Enable:      true,
	}
	er := c.naming.RegisterInstance(req)
	if er != nil {
		return errors.Wrap(er, " register service")
	}
	if c.Ephemeral {
		h := beat.NewHeartBeatService(c.naming.HttpClient(), req)
		go h.Start(c.stopC)
		c.heartBeatService = h
	}
	c.instance = req
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
