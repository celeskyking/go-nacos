package discovery

import (
	"github.com/pkg/errors"
	"gitlab.mfwdev.com/portal/go-nacos/api"
	"gitlab.mfwdev.com/portal/go-nacos/naming"
	beat "gitlab.mfwdev.com/portal/go-nacos/naming/heartbeat"
	"gitlab.mfwdev.com/portal/go-nacos/pkg/util"
	"gitlab.mfwdev.com/portal/go-nacos/types"
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
	Group string
	//心跳服务
	heartBeatService beat.HeartBeatService

	naming naming.NamingService

	Ephemeral bool

	instance *types.ServiceInstance
	//初始化的健康状态
	InitHealthy bool

	stopC chan struct{}
}

func NewDiscoveryClient(naming naming.NamingService, config *api.DiscoveryOptions) *Client {
	ip := util.LocalIP()
	c := &Client{
		IP:        ip,
		Port:      config.Port,
		Namespace: config.Namespace,
		Cluster:   config.Cluster,
		AppName:   config.AppName,
		Group:     config.Group,
		stopC:     make(chan struct{}, 0),
		naming:    naming,
		Ephemeral: true,
	}
	return c
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
		Group:     c.Group,
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
		GroupName:   c.Group,
		ClusterName: c.Cluster,
		Ephemeral:   c.Ephemeral,
		Weight:      1.0,
		Healthy:     c.InitHealthy,
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

func (c *Client) GetInstances(serviceName string, options *naming.QueryOptions) (*naming.ServerList, error) {
	return c.naming.GetInstances(serviceName, options)
}
