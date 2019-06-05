package api

import "time"

const (
	EndpointSwitch string = "ALIBABA_ALIWARE_ENDPOINT_PORT"

	Endpoint string = "ENDPOINT"

	EndpointPort string = "ENDPOINT_PORT"

	EndpointURL string = "ALIBABA_ALIWARE_ENDPOINT_PORT"
)

type HttpConfigOption struct {
	//连接超时
	ConnectTimeout time.Duration
	//当前nacos的api版本
	Version string
	//服务列表
	Servers []string

	LBStrategy LBStrategy
	//Endpoint 一个扩展的url,用来更新刷新远程服务器的servers列表的
	Endpoint string
	//EndpointEnabled 功能是否启动
	EndpointEnabled bool
}

type LBStrategy int

const (
	Direct LBStrategy = iota

	RoundRobin

	DefaultConnectTimeout = 5 * time.Second
)

type ConfigOption struct {
	//应用名称
	AppName string
	//环境名称
	Env string
	// 当前的namespace
	Namespace string
	//nacos 配置中心的地址
	Addresses []string
	//负载均衡策略
	LBStrategy LBStrategy
	//快照目录
	SnapshotDir string
	//当前服务的端口号
	Port int
	//Endpoint
	Endpoint string
	//EndpointEnabled 功能是否启动
	EndpointEnabled bool
}

func DefaultOption() *HttpConfigOption {
	return &HttpConfigOption{
		Version:        "v1",
		ConnectTimeout: DefaultConnectTimeout,
	}
}
