package naming

import (
	v1 "github.com/celeskyking/go-nacos/api/ns/v1"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"github.com/sirupsen/logrus"
	"math"
	"strings"
	"time"
)

const (
	Splitter           string = "@@"
	DefaultCacheMillis int    = 10
)

func NewServerList(httpClient v1.NamingHttpClient, receiver *v1.PushReceiver, watch bool, serviceName, groupName, namespaceId, clusters string) *ServerList {
	sl := &ServerList{
		receiver:    receiver,
		lb:          NewRandom(),
		watch:       watch,
		httpClient:  httpClient,
		ServiceName: serviceName,
		GroupName:   groupName,
		Clusters:    clusters,
		stopC:       make(chan struct{}, 0),
		LastRefTime: math.MaxInt64,
		NamespaceId: namespaceId,
		CacheMillis: DefaultCacheMillis,
	}
	return sl
}

type ServerList struct {
	receiver *v1.PushReceiver

	lb InternalLB

	watch bool
	//服务列表
	list []*types.ServiceInstance
	//最后更新的时间
	LastRefTime int64
	//缓存时间
	CacheMillis int
	// serviceName + clusters
	ServiceName string

	GroupName string

	NamespaceId string

	Clusters string

	httpClient v1.NamingHttpClient
	//缓存时间
	stopC chan struct{}
}

func (s *ServerList) GetAll() []*types.ServiceInstance {
	return s.lb.GetAll()
}

func (s *ServerList) Listen(stop <-chan struct{}) error {
	instances, result, er := selectInstances(s.httpClient, s.ServiceName, &QueryOptions{
		Group:     s.GroupName,
		Cluster:   s.Clusters,
		Namespace: s.NamespaceId,
		Watch:     true,
	})
	if er != nil {
		return er
	}
	s.CacheMillis = result.CacheMillis
	s.LastRefTime = result.LastRefTime
	s.lb.Refresh(instances)
	go func() {
		timer := time.NewTimer(time.Duration(s.CacheMillis) * time.Millisecond)
		for {
			select {
			case <-timer.C:
				instances, result, er := selectInstances(s.httpClient, s.ServiceName, &QueryOptions{
					Group:     s.GroupName,
					Cluster:   s.Clusters,
					Namespace: s.NamespaceId,
					Watch:     true,
				})
				if er != nil {
					logrus.Errorf("list service failed, serviceName:%s, error:%+v\n", result, er)
					time.Sleep(5 * time.Second)
					timer.Reset(time.Duration(s.CacheMillis) * time.Millisecond)
					continue
				}
				s.CacheMillis = result.CacheMillis
				s.LastRefTime = result.LastRefTime
				timer.Reset(time.Duration(s.CacheMillis) * time.Millisecond)
				s.lb.Refresh(instances)
			case <-s.stopC:
				return
			}
		}
	}()
	go func() {
		if s.watch {
			notifyC := s.receiver.GetNotifyChannel()
			for {
				select {
				case msg := <-notifyC:
					if Key(msg.Name, msg.Clusters) == Key(s.GroupName, s.ServiceName, s.Clusters) && msg.LastRefTime > s.LastRefTime {
						s.refreshServiceList(msg)
					}
				case <-s.stopC:
					return
				}
			}
		}
	}()
	return nil
}

func (s *ServerList) StopListen() {
	if s.watch {
		s.stopC <- struct{}{}
	}
}

func Key(parts ...string) string {
	return strings.Join(parts, Splitter)
}

func (s *ServerList) refreshServiceList(msg *v1.PushMessage) {
	var list []*types.ServiceInstance
	groupName := util.GetGroupName(msg.Name)
	for _, instance := range msg.Hosts {
		list = append(list, hostToServiceInstance(s.NamespaceId, groupName, instance))
	}
	s.list = list
	logrus.Info(util.ToJSONString(s.list))
	s.lb.Refresh(list)
}

func (s *ServerList) SelectOne() *types.ServiceInstance {
	return s.lb.SelectOne(nil)
}

func hostToServiceInstance(namespaceID, groupName string, msg *types.Host) *types.ServiceInstance {
	return &types.ServiceInstance{
		IP: msg.IP,

		Port: msg.Port,

		NamespaceID: namespaceID,

		Weight: msg.Weight,
		//是否上线
		Enable: msg.Enabled,
		//健康状态
		Healthy: msg.Healthy,
		//元数据
		Metadata: util.MapToString(msg.Metadata),
		//集群名
		ClusterName: msg.ClusterName,
		//服务名
		ServiceName: msg.ServiceName,

		GroupName: groupName,
		//临时节点
		Ephemeral: msg.Ephemeral,
	}
}
