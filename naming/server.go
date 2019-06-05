package naming

import (
	v1 "github.com/celeskyking/go-nacos/api/ns/v1"
	"github.com/celeskyking/go-nacos/types"
	"strings"
)

const (
	Splitter string = "@@"
)

func NewServerList(list []*types.ServiceInstance, receiver *v1.PushReceiver, watch bool, lastRefTime int64, serviceName, clusters string) *ServerList {
	sl := &ServerList{
		receiver:    receiver,
		lb:          NewRandom(),
		watch:       watch,
		list:        list,
		LastRefTime: lastRefTime,
		Key:         Key(serviceName, clusters),
		stopC:       make(chan struct{}, 0),
	}
	sl.lb.Refresh(list)
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

	// serviceName + clusters
	Key string

	stopC chan struct{}
}

func (s *ServerList) Start(stop <-chan struct{}) {
	go func() {
		if s.watch {
			notifyC := s.receiver.GetNotifyChannel()
			for {
				select {
				case msg := <-notifyC:
					if Key(msg.Name, msg.Clusters) == s.Key && msg.LastRefTime > s.LastRefTime {
						s.refreshServiceList(msg)
					}
				case <-s.stopC:
					return
				}
			}
		}
	}()
}

func (s *ServerList) Clear() {
	if s.watch {
		s.stopC <- struct{}{}
	}
}

func Key(serviceName, clusters string) string {
	parts := []string{serviceName, clusters}
	return strings.Join(parts, Splitter)
}

func (s *ServerList) refreshServiceList(msg *v1.PushMessage) {
	var list []*types.ServiceInstance
	for _, instance := range msg.Hosts {
		list = append(list, instance)
	}
	s.list = list
	s.lb.Refresh(list)
}

func (s *ServerList) SelectOne() *types.ServiceInstance {
	return s.lb.SelectOne(nil)
}
