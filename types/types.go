package types

import (
	"gitlab.mfwdev.com/portal/go-nacos/err"
	"strings"
)

const (
	FieldSeparator   byte = 2
	ArticleSeparator byte = 1
)

type ConfigsRequest struct {
	Tenant string `query:"tenant"`

	DataID string `query:"dataId"`

	Group string `query:"group"`
}

type ConfigsResponse struct {
	Value string
}

type ListenConfigsRequest struct {
	//监听的配置信息
	ListeningConfigs []*ListenKey
}

func (lcs *ListenConfigsRequest) Line() string {
	var result string
	for _, k := range lcs.ListeningConfigs {
		result += k.Line()
	}
	return result
}

type ListenChange struct {
	Key *ListenKey

	//新值
	NewValue string
}

type ListenKey struct {
	DataID string

	Group string

	ContentMD5 string

	Tenant string
}

func (l *ListenKey) Line() string {
	var result []string
	result = append(result, l.DataID)
	result = append(result, l.Group)
	result = append(result, l.ContentMD5)
	if l.Tenant != "" {
		result = append(result, l.Tenant)
	}
	return strings.Join(result, string(FieldSeparator)) + string(ArticleSeparator)
}

func ParseListenKey(line string) (*ListenKey, error) {
	line = strings.Trim(line, string(ArticleSeparator))
	parts := strings.Split(line, string(FieldSeparator))
	l := len(parts)
	if !(l == 3 || l == 4) {
		return nil, err.ErrKeyNotValid
	}
	key := &ListenKey{
		DataID: parts[0],
		Group:  parts[1],
	}
	if l == 3 {
		key.Tenant = parts[2]
	} else {
		key.ContentMD5 = parts[2]
		key.Tenant = parts[3]
	}
	return key, nil
}

func (l *ListenKey) ToConfigsRequest() *ConfigsRequest {
	return &ConfigsRequest{
		DataID: l.DataID,
		Group:  l.Group,
		Tenant: l.Tenant,
	}
}

type PublishConfig struct {
	Tenant string `query:"tenant"`

	DataID string `query:"dataId"`

	Group string `query:"group"`

	Content string `query:"content"`
}

type Result struct {
	Success bool
}

type ServiceInstance struct {
	IP string `query:"ip" validate:"required"`

	Port int `query:"port" validate:"required"`

	NamespaceID string `query:"namespaceId"`

	Weight float64 `query:"weight"`
	//是否上线
	Enable bool `query:"enable"`
	//健康状态
	Healthy bool `query:"healthy"`
	//元数据
	Metadata string `query:"metadata"`
	//集群名
	ClusterName string `query:"clusterName"`
	//服务名
	ServiceName string `query:"serviceName" validate:"required"`

	GroupName string `query:"groupName"`
	//临时节点
	Ephemeral bool `query:"ephemeral"`
}

type HeartBeat struct {
	ServiceName string `query:"serviceName"`

	GroupName string `query:"groupName"`

	NamespaceID string `query:"namespaceId"`
	//实例心跳内容
	Beat *Beat `query:"beat" transfer:"json"`
}

type Beat struct {
	ServiceName string `json:"serviceName"`

	IP string `json:"ip"`

	Port int `json:"port"`

	Cluster string `json:"cluster"`

	Weight float64 `json:"weight"`

	Metadata map[string]string `json:"metadata,omitempty"`
}

type HeartBeatResult struct {

	//间隔毫秒数
	ClientBeatInterval int `json:"clientBeatInterval"`
}

type ServiceInstanceListOption struct {
	ServiceName string `query:"serviceName" validate:"required"`

	NamespaceID string `query:"namespaceId"`

	Clusters string `query:"clusters"`
	//默认为false
	HealthyOnly bool `query:"healthyOnly"`

	UdpPort int `query:"udpPort"`

	ClientIP string `query:"clientIP"`
}

type ServiceInstanceListResult struct {
	Dom string `json:"dom"`

	Name string `json:"name"`

	CacheMillis int `json:"cacheMillis"`

	Hosts []*Host `json:"hosts"`

	CheckSum string `json:"checkSum"`

	LastRefTime int64 `json:"lastRefTime"`

	Env string `json:"env"`

	Clusters string `json:"clusters"`

	Metadata map[string]string `json:"metadata"`
}

type Host struct {
	Valid bool `json:"valid"`

	Marked bool `json:"marked"`

	InstanceID string `json:"instanceId"`

	Enabled bool `json:"enabled"`

	Healthy bool `json:"healthy"`

	Port int `json:"port"`

	IP string `json:"ip"`

	Weight float64 `json:"weight"`

	Metadata map[string]string `json:"metadata,omitempty"`

	ServiceName string `json:"serviceName"`

	Ephemeral bool `json:"ephemeral"`

	ClusterName string `json:"clusterName"`
}

type InstanceDetail struct {
	Metadata map[string]string `json:"metadata,omitempty"`

	InstanceID string `json:"instanceId"`

	Port int `json:"port"`

	Service string `json:"service"`

	Healthy bool `json:"healthy"`

	Enabled bool `json:"enabled"`

	IP string `json:"ip"`

	ClusterName string `json:"clusterName"`

	Weight float64 `json:"weight"`
}

type Service struct {
	ServiceName string `query:"serviceName" validate:"required"`

	GroupName string `query:"groupName"`

	NamespaceId string `query:"namespaceId"`
	//保护阈值
	ProtectThreshold float64 `query:"protectThreshold"`

	Metadata string `query:"metadata"`
	//访问策略
	Selector string `query:"selector"`
}

//查询
type ServiceDetail struct {
	Metadata map[string]string `json:"metadata"`

	GroupName string `json:"groupName"`

	NamespaceID string `json:"namespaceId"`

	Name string `json:"name"`

	Selector map[string]string `json:"selector,omitempty"`

	ProtectThreshold float64 `json:"protectThreshold"`

	Clusters []*ClusterDetail `json:"clusters"`
}

type ClusterDetail struct {
	HealthChecker map[string]interface{} `json:"healthChecker"`

	Metadata map[string]string `json:"metadata,omitempty"`

	Name string `json:"name"`
}

type ServiceListOption struct {
	PageNo int `query:"pageNo" validate:"required"`

	PageSize int `query:"pageSize" validate:"required"`

	GroupName string `query:"groupName"`

	NamespaceID string `query:"namespaceId"`
}

type CatalogServiceDetail struct {
	ServiceName string `json:"serviceName"`

	GroupName string `json:"groupName"`

	ClusterMap map[string]*ClusterInfo `json:"clusterMap"`

	Metadata map[string]string `json:"metadata"`
}

type ClusterInfo struct {
	Hosts []*IPAddressInfo `json:"hosts"`
}

type IPAddressInfo struct {
	Valid bool `json:"valid"`

	Metadata map[string]string `json:"metadata,omitempty"`

	Port int `json:"port"`

	IP string `json:"ip"`

	Weight float64 `json:"weight"`

	Enabled bool `json:"enabled"`
}

type ServiceListResult struct {
	Count int `json:"count"`

	Doms []string `json:"doms"`
}

//SwitchesDetail 开关的详情
type SwitchesDetail struct {
	Name string `json:"name"`

	Masters []string `json:"masters"`

	AddWeightMap map[string]string `json:"addWeightMap"`

	DefaultPushCacheMillis int `json:"defaultPushCacheMillis"`

	DistroThreshold float64 `json:"distroThreshold"`

	HealthCheckEnabled bool `json:"healthCheckEnabled"`

	DistroEnabled bool `json:"distroEnabled"`

	EnableStandalone bool `json:"enableStandalone"`

	PushEnabled bool `json:"pushEnabled"`

	CheckTimes int `json:"checkTimes"`

	HttpHealthParams HealthCheckParams `json:"httpHealthParams"`

	TcpHealthParams HealthCheckParams `json:"tcpHealthParams"`

	MySQLHealthParams HealthCheckParams `json:"mysqlHealthParams"`

	IncrementalList []string `json:"incrementalList"`

	ServerStatusSynchronizationPeriodMillis int `json:"serverStatusSynchronizationPeriodMillis"`

	ServiceStatusSynchronizationPeriodMillis int `json:"serviceStatusSynchronizationPeriodMillis"`

	DisableAddIP bool `json:"disableAddIP"`

	SendBeatOnly bool `json:"sendBeatOnly"`

	LimitedUrlMap map[string]int `json:"limitedUrlMap,omitempty"`

	DistroServerExpiredMillis int `json:"distroServerExpiredMillis"`

	PushGoVersion string `json:"pushGoVersion"`

	PushJavaVersion string `json:"pushJavaVersion"`

	PushPythonVersion string `json:"pushPythonVersion"`

	PushCVersion string `json:"pushCVersion"`

	EnableAuthentication bool `json:"enableAuthentication"`

	OverriddenServerStatus string `json:"overriddenServerStatus"`

	DefaultInstanceEphemeral bool `json:"defaultInstanceEphemeral"`

	HealthCheckWhiteList []string `json:"healthCheckWhiteList"`

	Checksum string `json:"checksum"`
}

type HealthCheckParams struct {
	Max int `json:"max"`

	Min int `json:"min"`

	Factor float64 `json:"factor"`
}

type UpdateSwitchRequest struct {

	//开关名
	Entry string `query:"entry" validate:"required"`

	//开关之
	Value string `query:"value" validate:"required"`
}

type NacosServers struct {
	Servers []*NacosServer `json:"servers"`
}

type NacosServer struct {
	IP string `json:"ip"`

	ServePort int `json:"servePort"`

	Site string `json:"site"`

	Weight int `json:"weight"`

	AdWeight int `json:"ad_weight"`

	Alive bool `json:"alive"`

	LastRefTime int `json:"lastRefTime"`

	LastRefTimeStr string `json:"lastRefTimeStr"`
}

type NacosLeader struct {
	HeartbeatDueMs int64 `json:"heartbeatDueMs"`

	LeaderDueMs int64 `json:"leaderDueMs"`

	Term int64 `json:"term"`

	Ip string `json:"ip"`

	VoteFor string `json:"voteFor"`

	//LEADER, FOLLOWER, CANDIDATE
	State string `json:"state"`
}

type UpdateServiceInstanceHealthyRequest struct {
	IP string `query:"ip" validate:"required"`

	Port int `query:"port" validate:"required"`

	NamespaceID string `query:"namespaceId"`

	//健康状态
	Healthy bool `query:"healthy"`
	//集群名
	ClusterName string `query:"clusterName"`
	//服务名
	ServiceName string `query:"serviceName" validate:"required"`

	GroupName string `query:"groupName"`
}

type Metrics struct {
	ServerCount int `json:"serverCount"`

	Load float64 `json:"load"`

	Mem float64 `json:"mem"`

	ResponsibleServiceCount int `json:"responsibleServiceCount"`

	InstanceCount int `json:"instanceCount"`

	Cpu float64 `json:"cpu"`

	Status string `json:"status"`
	//
	ResponsibleInstanceCount int `json:"responsibleInstanceCount"`
}

type Cluster struct {
	NamespaceID string `query:"namespaceId" json:"namespaceId"`

	ClusterName string `query:"clusterName" json:"clusterName"`

	ServiceName string `query:"serviceName" json:"serviceName"`

	GroupName string `query:"groupName" json:"groupName"`

	Metadata map[string]string `query:"metadata" json:"metadata,omitempty" transfer:"json"`

	CheckPort int `query:"checkPort" json:"checkPort"`

	HealthChecker IHealthChecker `query:"healthChecker" json:"healthChecker" transfer:"json"`

	UseInstancePort4Check bool `query:"useInstancePort4Check" json:"-"`
}

type HealthChecker struct {
	Type string `json:"type"`
}

type IHealthChecker interface {
	GetType() string
}

type MySQLHealthChecker struct {
	Type string `json:"type"`

	User string `json:"user"`

	Pwd string `json:"pwd"`

	Cmd string `json:"cmd"`
}

func (m *MySQLHealthChecker) GetType() string {
	return m.Type
}

type TCPHealthChecker struct {
	Type string `json:"type"`
}

func (t *TCPHealthChecker) GetType() string {
	return t.Type
}

type HttpHealthChecker struct {
	Type string `json:"type"`

	Path string `json:"path"`

	Headers string `json:"headers"`

	ExpectedResponseCode int `json:"expectedResponseCode"`
}

func (h *HttpHealthChecker) GetType() string {
	return h.Type
}

func (h *HttpHealthChecker) SetExpectedResponseCode(code int) {
	h.ExpectedResponseCode = code
}

func (h *HttpHealthChecker) SetHeaders(headers string) {
	h.Headers = headers
}

type NoneHealthChecker struct {
	Type string `json:"type"`
}

func (n *NoneHealthChecker) GetType() string {
	return n.Type
}

func NewNoneHealthChecker() *NoneHealthChecker {
	return &NoneHealthChecker{
		Type: "NONE",
	}
}

func NewHttpHealthChecker(path string) *HttpHealthChecker {
	return &HttpHealthChecker{
		Path: path,
	}
}

func NewTcpHealthChecker() *TCPHealthChecker {
	return &TCPHealthChecker{
		Type: "TCP",
	}
}

func NewMySQLHealthChecker(user, pwd, cmd string) *MySQLHealthChecker {
	return &MySQLHealthChecker{
		Type: "MYSQL",
		User: user,
		Pwd:  pwd,
		Cmd:  cmd,
	}
}

type FileDesc struct {
	//
	Name string
	//应用名
	AppName string
	//环境
	Group string

	Namespace string
}
