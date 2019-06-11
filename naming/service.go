package naming

import (
	"github.com/pkg/errors"
	"gitlab.mfwdev.com/portal/go-nacos/api"
	v1 "gitlab.mfwdev.com/portal/go-nacos/api/ns/v1"
	"gitlab.mfwdev.com/portal/go-nacos/err"
	"gitlab.mfwdev.com/portal/go-nacos/pkg/util"
	"gitlab.mfwdev.com/portal/go-nacos/types"
	"strconv"
	"strings"
	"sync"
)

//Service naming的服务
type NamingService interface {

	//注册实例
	RegisterInstance(serviceInstance *types.ServiceInstance) error

	//注销
	DeRegisterInstance(serviceInstance *types.ServiceInstance) error

	//GetInstances 通过服务名来获取实例信息,默认会选择同namespace和group下的服务
	GetInstances(serviceName string, options *QueryOptions) (*ServerList, error)

	//GetAllService 返回所有的服务信息,服务的个数和服务的名称列表
	GetAllServices(namespaceID string) (int, []*types.CatalogServiceDetail, error)

	//返回所有的
	GetAllServicesCount(namespaceID string) (int, error)

	//更新cluster
	PatchCluster(cluster *types.Cluster) error

	//设置实例是否可用
	SetInstanceHealthy(serviceName string, options *QueryOptions) error

	//Stop 停止当前的服务
	Stop()

	//http客户端
	HttpClient() v1.NamingHttpClient

	//Config
	GetConfig() *api.ServerOptions
}

//QueryOptions 查询的选项
type QueryOptions struct {
	//命名空间
	Namespace string
	//环境
	Group string
	//集群名
	Cluster string
	//是否健康
	Healthy bool
	//是否开启watch
	Watch bool
}

func NewNamingService(config *api.ServerOptions) NamingService {
	httpOption := api.DefaultOption()
	httpOption.Servers = config.Addresses
	httpOption.LBStrategy = config.LBStrategy
	httpClient := v1.NewNamingHttpClient(httpOption)
	stopC := make(chan struct{})
	ns := &namingService{
		Config:       config,
		httpClient:   httpClient,
		pushReceiver: v1.NewPushReceiver(),
		stopC:        stopC,
	}
	go ns.pushReceiver.Start()
	return ns
}

// namingService implement NamingService
type namingService struct {
	//当前的配置信息
	Config *api.ServerOptions

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

func (n *namingService) SetInstanceHealthy(serviceName string, option *QueryOptions) error {
	r, er := n.httpClient.UpdateServiceInstance(&types.ServiceInstance{
		NamespaceID: option.Namespace,
		ServiceName: serviceName,
		GroupName:   option.Group,
		ClusterName: option.Cluster,
		Healthy:     option.Healthy,
	})
	if er != nil {
		return er
	}
	if !r.Success {
		return err.ErrNamingService
	}
	return nil
}

func (n *namingService) GetConfig() *api.ServerOptions {
	return n.Config
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

func (n *namingService) PatchCluster(cluster *types.Cluster) error {
	r, er := n.httpClient.PatchCluster(cluster)
	if er != nil {
		return errors.Wrap(er, "patch cluster instance")
	}
	if !r.Success {
		return err.ErrNamingService
	}
	return nil
}

//GetInstances 返回制定services的所有的实例信息
//todo 支持failover
func (n *namingService) GetInstances(serviceName string, options *QueryOptions) (*ServerList, error) {
	sl := NewServerList(n.httpClient, n.pushReceiver, options.Watch, serviceName, options.Group, options.Namespace, options.Cluster)
	er := sl.Listen(n.stopC)
	return sl, er
}

func (n *namingService) GetAllServices(namespaceID string) (int, []*types.CatalogServiceDetail, error) {
	services, er := n.httpClient.CatalogServices(true, namespaceID)
	if er != nil {
		return 0, nil, er
	}
	return len(services), services, nil
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

func selectInstances(httpClient v1.NamingHttpClient, serviceName string, options *QueryOptions) (instances []*types.ServiceInstance, result *types.ServiceInstanceListResult, er error) {
	req := buildQueryListRequest(serviceName, options, options.Watch)
	r, er := httpClient.ListServiceInstance(req)
	if er != nil {
		return nil, nil, er
	}
	if len(r.Hosts) == 0 {
		return nil, nil, nil
	}
	var results []*types.ServiceInstance
	for _, h := range r.Hosts {
		parts := strings.Split(h.InstanceID, "#")
		results = append(results, &types.ServiceInstance{
			IP:          parts[0],
			Port:        stringToInt(parts[1]),
			ServiceName: parts[3],
			GroupName:   options.Group,
			NamespaceID: options.Namespace,
			ClusterName: h.ClusterName,
			Ephemeral:   h.Ephemeral,
			Metadata:    util.MapToString(h.Metadata),
		})
	}
	return results, r, nil
}

func buildQueryListRequest(appName string, options *QueryOptions, subscriber bool) *types.ServiceInstanceListOption {
	req := &types.ServiceInstanceListOption{}
	req.ServiceName = appName
	if options.Group != "" {
		req.ServiceName = options.Group + Splitter + appName
	}
	if options.Namespace != "" {
		req.NamespaceID = options.Namespace
	}
	if len(options.Cluster) != 0 {
		req.Clusters = options.Cluster
	}
	req.HealthyOnly = options.Healthy
	if subscriber {
		req.ClientIP = util.LocalIP()
		req.UdpPort = v1.GetUDPPort()
	}
	return req
}

func stringToInt(port string) int {
	r, _ := strconv.Atoi(port)
	return r
}
