package service

import (
	"github.com/celeskyking/go-nacos/api"
	v1 "github.com/celeskyking/go-nacos/api/ns/v1"
	"github.com/celeskyking/go-nacos/err"
	"github.com/celeskyking/go-nacos/naming"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"sync"
)

//Service naming的服务
type NamingService interface {

	//注册实例
	RegisterInstance(serviceInstance *types.ServiceInstance) error

	DeRegisterInstance(instance *types.ServiceInstance) error

	//GetInstances 通过服务名来获取实例信息,默认会选择同namespace和group下的服务
	GetInstances(serviceName string, options *QueryOptions) (*naming.ServerList, error)

	//GetAllService 返回所有的服务信息,服务的个数和服务的名称列表
	GetAllServices(namespaceID string) (int, []string, error)

	//返回命名空间下制定页的所有服务
	GetServices(namespaceID string, pageNo, pageSize int) (int, []string, error)

	//返回所有的
	GetAllServicesCount(namespaceID string) (int, error)

	//Stop 停止当前的服务
	Stop()

	//http客户端
	HttpClient() v1.NamingHttpClient
}

//QueryOptions 查询的选项
type QueryOptions struct {
	//命名空间
	Namespace string
	//环境
	Group string
	//集群名
	Cluster []string
	//是否健康
	Healthy bool
	//是否开启watch
	Watch bool
}

func NewNamingService(config *api.ConfigOption) NamingService {
	httpOption := api.DefaultOption()
	httpOption.Servers = config.Addresses
	httpOption.LBStrategy = config.LBStrategy
	httpClient := v1.NewNamingHttpClient(httpOption)
	lb := httpClient.LoadBalance()
	stopC := make(chan struct{})
	ns := &namingService{
		Config:       config,
		httpClient:   httpClient,
		pushReceiver: v1.NewPushReceiver(lb),
		stopC:        stopC,
	}
	return ns
}

// namingService implement NamingService
type namingService struct {
	//当前的配置信息
	Config *api.ConfigOption

	httpClient v1.NamingHttpClient
	//推送接收
	pushReceiver *v1.PushReceiver

	lock sync.Mutex

	stopC chan struct{}
}

func (n *namingService) HttpClient() v1.NamingHttpClient {
	return n.httpClient
}

func (n *namingService) RegisterInstance(serviceInstance *types.ServiceInstance) error {
	r, er := n.httpClient.RegisterServiceInstance(serviceInstance)
	if er != nil {
		return errors.Wrap(er, "register service instance")
	}
	if !r.Success {
		return err.ErrNamingService
	}
	return nil
}

func (n *namingService) DeRegisterInstance(serviceInstance *types.ServiceInstance) error {
	r, er := n.httpClient.DeRegisterServiceInstance(serviceInstance)
	if er != nil {
		return errors.Wrap(er, "deRegister service instance")
	}
	if !r.Success {
		return err.ErrNamingService
	}
	return nil
}

//GetInstances 返回制定services的所有的实例信息
//todo 支持failover
func (n *namingService) GetInstances(serviceName string, options *QueryOptions) (*naming.ServerList, error) {
	r, lastRefTime, clusters, er := n.SelectInstances(serviceName, options)
	if er != nil {
		return nil, er
	}
	sl := naming.NewServerList(r, n.pushReceiver, options.Watch, lastRefTime, serviceName, clusters)
	sl.Start(n.stopC)
	return sl, nil
}

func (n *namingService) GetAllServices(namespaceID string) (int, []string, error) {
	c, er := n.GetAllServicesCount(namespaceID)
	if er != nil {
		return 0, nil, er
	}
	r, er := n.httpClient.ListService(&types.ServiceListOption{PageNo: 1, PageSize: c, NamespaceID: namespaceID})
	if er != nil {
		return 0, nil, er
	}
	return r.Count, r.Doms, nil
}

func (n *namingService) GetServices(namespaceID string, pageNo, pageSize int) (int, []string, error) {
	r, er := n.httpClient.ListService(&types.ServiceListOption{PageNo: pageNo, PageSize: pageSize, NamespaceID: namespaceID})
	if er != nil {
		return 0, nil, er
	}
	return r.Count, r.Doms, nil
}

func (n *namingService) GetAllServicesCount(namespaceID string) (int, error) {
	r, er := n.httpClient.ListService(&types.ServiceListOption{PageNo: 1, PageSize: 1, NamespaceID: namespaceID})
	if er != nil {
		return 0, er
	}
	return r.Count, nil
}

func (n *namingService) Stop() {
	select {
	case n.stopC <- struct{}{}:
	default:

	}

}

func (n *namingService) SelectInstances(serviceName string, options *QueryOptions) (instances []*types.ServiceInstance, lastRefTime int64, clusters string, er error) {
	req := buildQueryListRequest(serviceName, options, false)
	r, er := n.httpClient.ListServiceInstance(req)
	if er != nil {
		return nil, 0, "", er
	}
	if len(r.Hosts) == 0 {
		return nil, 0, "", nil
	}
	var results []*types.ServiceInstance
	for _, h := range r.Hosts {
		parts := strings.Split(h.InstanceID, "-")
		results = append(results, &types.ServiceInstance{
			IP:          parts[0],
			Port:        stringToInt(parts[1]),
			ServiceName: parts[3],
			GroupName:   options.Group,
			NamespaceID: options.Namespace,
			ClusterName: parts[2],
			Ephemeral:   true,
			Metadata:    util.MapToString(h.Metadata),
		})
	}
	return results, r.LastRefTime, r.Clusters, nil
}

func buildQueryListRequest(appName string, options *QueryOptions, subscriber bool) *types.ServiceInstanceListOption {
	req := &types.ServiceInstanceListOption{}
	req.ServiceName = appName
	if options.Group != "" {
		req.GroupName = options.Group
	}
	if options.Namespace != "" {
		req.NamespaceID = options.Namespace
	}
	if len(options.Cluster) != 0 {
		req.Clusters = strings.Join(options.Cluster, ",")
	}
	req.HealthyOnly = options.Healthy
	if subscriber {
		req.ClientIP = util.LocalIP()
		req.UdpPort = v1.UDPPort
	}
	return req
}

func stringToInt(port string) int {
	r, _ := strconv.Atoi(port)
	return r
}
