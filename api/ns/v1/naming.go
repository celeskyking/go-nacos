package v1

import (
	"encoding/json"
	"fmt"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/api/ns/endpoint"
	"github.com/celeskyking/go-nacos/client/http"
	"github.com/celeskyking/go-nacos/client/loadbalancer"
	"github.com/celeskyking/go-nacos/err"
	"github.com/celeskyking/go-nacos/pkg/query"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"
)

const (
	Prefix                = "nacos"
	DefaultConnectTimeout = 5 * time.Second
	InstancePath          = "/ns/instance"
	InstanceListPath      = "/ns/instance/list"
	InstanceHeartBeatPath = "/ns/instance/beat"
	ServicePath           = "/ns/service"
	ServiceListPath       = "/ns/service/list"
	SwitchesPath          = "/ns/operator/switches"
	MetricsPath           = "/ns/operator/metrics"
	NacosServersPath      = "/ns/operator/servers"
	LeaderPath            = "/ns/raft/leader"
	InstanceHealthPath    = "/ns/health/instance"
	CatalogServicesPath   = "/ns/catalog/services"
	HealthPath            = "/v1/console/health/liveness"
)

type NamingHttpClient interface {
	//RegisterServiceInstance 注册服务
	RegisterServiceInstance(instance *types.ServiceInstance) (*types.Result, error)

	DeRegisterServiceInstance(instance *types.ServiceInstance) (*types.Result, error)

	UpdateServiceInstance(instance *types.ServiceInstance) (*types.Result, error)

	GetServiceInstanceDetail(instance *types.ServiceInstance) (*types.InstanceDetail, error)

	ListServiceInstance(option *types.ServiceInstanceListOption) (*types.ServiceInstanceListResult, error)

	HeartBeat(beat *types.HeartBeat) (*types.HeartBeatResult, error)

	CreateService(service *types.Service) (*types.Result, error)

	DeleteService(service *types.Service) (*types.Result, error)

	UpdateService(service *types.Service) (*types.Result, error)
	//Service
	GetService(service *types.Service) (*types.ServiceDetail, error)
	//服务列表
	//必须严格匹配groupName和namespace
	ListService(option *types.ServiceListOption) (*types.ServiceListResult, error)

	GetSwitches() (*types.SwitchesDetail, error)

	GetMetrics() (*types.Metrics, error)

	GetNacosServers() (*types.NacosServers, error)

	GetLeader() (*types.NacosLeader, error)

	UpdateSwitches(request *types.UpdateSwitchRequest) (*types.Result, error)

	UpdateServiceInstanceHealthy(request *types.UpdateServiceInstanceHealthyRequest) (*types.Result, error)

	CatalogServices(withInstances bool, namespaceId string) ([]*types.CatalogServiceDetail, error)

	LoadBalance() loadbalancer.LB

	Stop()
}

func NewNamingHttpClient(option *api.HttpConfigOption) NamingHttpClient {
	ss := option.Servers
	if len(ss) == 0 {
		logrus.Errorf("不合法的nacos服务器列表,服务器最少存在一个")
		os.Exit(1)
	}
	var servers []*loadbalancer.Server
	for _, s := range ss {
		u, er := api.ToURL(s)
		if er != nil {
			logrus.Errorf("不合法的server地址:%s", s)
			os.Exit(1)
		}
		server := loadbalancer.NewServer(u, 100, path.Join(Prefix, HealthPath))
		servers = append(servers, server)
	}
	var e *endpoint.Endpoint
	var serverChanges chan []string
	stopC := make(chan struct{})
	if option.EndpointEnabled {
		endpointAddr := option.Endpoint
		if endpointAddr == "" {
			logrus.Fatal("endpoint的address不能够为空")
			os.Exit(1)
		}
		e = endpoint.NewEndpoint(endpointAddr)
		serverChanges = e.Run(stopC)
	}

	ch := &namingHttpClient{}
	if option.LBStrategy == api.RoundRobin {
		ch.LB = loadbalancer.NewRoundRobin(servers, true)
	} else {
		ch.LB = loadbalancer.NewDirectProxy(servers)
	}
	ch.Option = option
	go func() {
		for ss := range serverChanges {
			var servers []*loadbalancer.Server
			for _, s := range ss {
				u, er := api.ToURL(s)
				if er != nil {
					logrus.Errorf("不合法的server地址:%s", s)
					os.Exit(1)
				}
				server := loadbalancer.NewServer(u, 100, path.Join(Prefix, HealthPath))
				servers = append(servers, server)
			}
			if len(servers) > 0 {
				ch.LB.RefreshServers(servers)
			}
		}
	}()
	return ch
}

//namingHttpClient implement NamingHttpClient service
type namingHttpClient struct {
	LB loadbalancer.LB

	Option *api.HttpConfigOption

	endpoint *endpoint.Endpoint

	stopC chan struct{}
}

func (n *namingHttpClient) CatalogServices(withInstances bool, namespaceId string) ([]*types.CatalogServiceDetail, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, CatalogServicesPath) + "?namespaceId=" + namespaceId
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Get(url).EndBytes()
	er := handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	var servers []*types.CatalogServiceDetail
	er = json.Unmarshal(body, &servers)
	return servers, er
}

func (n *namingHttpClient) GetNacosServers() (*types.NacosServers, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, NacosServersPath)
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Get(url).EndBytes()
	er := handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	var servers types.NacosServers
	er = json.Unmarshal(body, &servers)
	return &servers, er
}

func (n *namingHttpClient) LoadBalance() loadbalancer.LB {
	return n.LB
}

func (n *namingHttpClient) Stop() {
	n.stopC <- struct{}{}
}

func (n *namingHttpClient) RegisterServiceInstance(instance *types.ServiceInstance) (*types.Result, error) {
	logrus.Infof("register service instance:%s", util.ToJSONString(instance))
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, InstancePath)
	req, er := query.Marshal(instance)
	if er != nil {
		return nil, er
	}
	fmt.Println("register instance:" + req)
	resp, body, errs := http.NewNamingHttp().Post(url).Query(req).End()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	return &types.Result{
		Success: body == "ok",
	}, nil
}

func (n *namingHttpClient) DeRegisterServiceInstance(instance *types.ServiceInstance) (*types.Result, error) {
	logrus.Infof("DeRegister service instance:%s", util.ToJSONString(instance))
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, InstancePath)
	req, er := query.Marshal(instance)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Delete(url).Query(req).End()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	return &types.Result{
		Success: body == "ok",
	}, nil
}

func (n *namingHttpClient) UpdateServiceInstance(instance *types.ServiceInstance) (*types.Result, error) {
	logrus.Infof("update service instance:%s", util.ToJSONString(instance))
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, InstancePath)
	req, er := query.Marshal(instance)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Put(url).Query(req).End()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	return &types.Result{
		Success: body == "ok",
	}, nil
}

func (n *namingHttpClient) GetServiceInstanceDetail(instance *types.ServiceInstance) (*types.InstanceDetail, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, InstancePath)
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Get(url).EndBytes()
	er := handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	var detail types.InstanceDetail
	er = json.Unmarshal(body, &detail)
	return &detail, er
}

func (n *namingHttpClient) ListServiceInstance(option *types.ServiceInstanceListOption) (*types.ServiceInstanceListResult, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, InstanceListPath)
	req, er := query.Marshal(option)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Get(url).Query(req).EndBytes()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	var detail types.ServiceInstanceListResult
	er = json.Unmarshal(body, &detail)
	return &detail, er
}

func (n *namingHttpClient) HeartBeat(beat *types.HeartBeat) (*types.HeartBeatResult, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, InstanceHeartBeatPath)
	req, er := query.Marshal(beat)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Put(url).Query(req).End()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	var r types.HeartBeatResult
	er = json.Unmarshal([]byte(body), &r)
	if er != nil {
		return nil, er
	}
	return &r, nil
}

func (n *namingHttpClient) CreateService(service *types.Service) (*types.Result, error) {
	logrus.Infof("create service:%s", util.ToJSONString(service))
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, ServicePath)
	req, er := query.Marshal(service)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Post(url).Query(req).End()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	return &types.Result{
		Success: body == "ok",
	}, nil
}

func (n *namingHttpClient) DeleteService(service *types.Service) (*types.Result, error) {
	logrus.Infof("delete service:%s", util.ToJSONString(service))
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, ServicePath)
	req, er := query.Marshal(service)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Delete(url).Query(req).End()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	return &types.Result{
		Success: body == "ok",
	}, nil
}

func (n *namingHttpClient) UpdateService(service *types.Service) (*types.Result, error) {
	logrus.Infof("delete service:%s", util.ToJSONString(service))
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, ServicePath)
	req, er := query.Marshal(service)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Put(url).Query(req).End()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	return &types.Result{
		Success: body == "ok",
	}, nil
}

func (n *namingHttpClient) GetService(service *types.Service) (*types.ServiceDetail, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, ServicePath)
	req, er := query.Marshal(service)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Get(url).Query(req).EndBytes()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	var detail types.ServiceDetail
	er = json.Unmarshal(body, &detail)
	return &detail, er
}

func (n *namingHttpClient) ListService(option *types.ServiceListOption) (*types.ServiceListResult, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, ServiceListPath)
	req, er := query.Marshal(option)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Get(url).Query(req).EndBytes()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	var list types.ServiceListResult
	er = json.Unmarshal(body, &list)
	return &list, er
}

func (n *namingHttpClient) GetSwitches() (*types.SwitchesDetail, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, SwitchesPath)
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Get(url).EndBytes()
	er := handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	var switches types.SwitchesDetail
	er = json.Unmarshal(body, &switches)
	return &switches, er
}

func (n *namingHttpClient) GetMetrics() (*types.Metrics, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, MetricsPath)
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Get(url).EndBytes()
	er := handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	var metrics types.Metrics
	er = json.Unmarshal(body, &metrics)
	return &metrics, er
}

func (n *namingHttpClient) GetLeader() (*types.NacosLeader, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, LeaderPath)
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Get(url).EndBytes()
	er := handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	var leader types.NacosLeader
	er = json.Unmarshal(body, &leader)
	return &leader, er
}

func (n *namingHttpClient) UpdateSwitches(request *types.UpdateSwitchRequest) (*types.Result, error) {
	logrus.Infof("update switches:%s", util.ToJSONString(request))
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, SwitchesPath)
	req, er := query.Marshal(request)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Put(url).Query(req).EndBytes()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	b := string(body) == "ok"
	return &types.Result{Success: b}, er
}

func (n *namingHttpClient) UpdateServiceInstanceHealthy(request *types.UpdateServiceInstanceHealthyRequest) (*types.Result, error) {
	url := api.SelectOne(n.LB) + path.Join(Prefix, n.Option.Version, InstanceHealthPath)
	req, er := query.Marshal(request)
	if er != nil {
		return nil, er
	}
	resp, body, errs := http.NewNamingHttp().Timeout(DefaultConnectTimeout).Put(url).Query(req).EndBytes()
	er = handleErrorResponse(resp, errs)
	if er != nil {
		return nil, er
	}
	b := string(body) == "ok"
	return &types.Result{Success: b}, er
}

func handleErrorResponse(resp gorequest.Response, errs []error) error {
	if resp != nil && errs != nil {
		data, er := ioutil.ReadAll(resp.Body)
		if er != nil {
			return errors.New("go-nacos system error,statusCode:5xx")
		}
		return err.NewHttpClientError(string(data), errs...)
	}
	if resp == nil {
		return err.NewHttpClientError("valid response")
	} else if resp.StatusCode == 200 {
		return nil
	} else {
		return err.NewHttpClientError("response status code valid,code:"+strconv.Itoa(resp.StatusCode), nil)
	}
}
