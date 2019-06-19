package api

import "time"

type HttpConfigOption struct {
	//连接超时
	ConnectTimeout time.Duration
	//当前nacos的api版本
	Version string
	//服务列表
	Servers []string
	//LBStrategy 负载均衡策略
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

type ServerOptions struct {
	//nacos 配置中心的地址
	Addresses []string
	//负载均衡策略
	LBStrategy LBStrategy
	//Endpoint
	Endpoint string
	//EndpointEnabled 功能是否启动
	EndpointEnabled bool
	//命名空间地址
	NamespaceID string
}

type AppConfig struct {
	//应用名称
	AppName string
	//环境名称
	Group string
	// 当前的namespace
	Namespace string
	//当前的实例集群
	Cluster string
	//当前服务的端口号
	Port int
	//当前机器的ip地址,可以为空或者设置LOCAL_IP的环境变量, 自动获取本地ip地址
	IP string
}

//ConfigOptions 把nacos的定义高级抽象了一下,把Group拆分为app和env group= app:env
type ConfigOptions struct {
	*ServerOptions

	SnapshotDir string
}

type DiscoveryOptions struct {
	//应用名称
	AppName string
	//环境名称
	Group string
	// 当前的namespace
	Namespace string
	//当前的实例集群
	Cluster string
	//当前服务的端口号
	Port int

	IP string
}

func DefaultOption() *HttpConfigOption {
	return &HttpConfigOption{
		Version:        "v1",
		ConnectTimeout: DefaultConnectTimeout,
	}
}
