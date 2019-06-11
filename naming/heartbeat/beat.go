package beat

import (
	"github.com/sirupsen/logrus"
	v1 "gitlab.mfwdev.com/portal/go-nacos/api/ns/v1"
	"gitlab.mfwdev.com/portal/go-nacos/pkg/util"
	"gitlab.mfwdev.com/portal/go-nacos/types"
	"time"
)

//HeartBeat
type HeartBeatService interface {

	//开始
	Start(stop <-chan struct{})
}

func NewHeartBeatService(client v1.NamingHttpClient, instance *types.ServiceInstance) HeartBeatService {
	return &heartBeatService{
		client:   client,
		Instance: instance,
	}
}

type heartBeatService struct {
	client v1.NamingHttpClient

	Instance *types.ServiceInstance
}

func (h *heartBeatService) Start(stop <-chan struct{}) {
	interval := h.beat()
	timer := time.NewTimer(time.Duration(interval) * time.Millisecond)
	for {
		select {
		case <-timer.C:
			//logrus.Info("heart beat")
			interval = h.beat()
			timer.Reset(time.Duration(interval) * time.Millisecond)
		case <-stop:
			timer.Stop()
			logrus.Infof("heartbeat stopped")
			return
		}
	}
}

func (h *heartBeatService) beat() int {
	reties := 0
	maxDelay := 30
	for {
		r, err := h.client.HeartBeat(&types.HeartBeat{
			ServiceName: h.Instance.ServiceName,
			NamespaceID: h.Instance.NamespaceID,
			GroupName:   h.Instance.GroupName,
			Beat: &types.Beat{
				ServiceName: h.Instance.ServiceName,
				IP:          h.Instance.IP,
				Port:        h.Instance.Port,
				Cluster:     h.Instance.ClusterName,
				Weight:      h.Instance.Weight,
				Metadata:    util.ToMap(h.Instance.Metadata),
			},
		})
		if err == nil {
			return r.ClientBeatInterval
		} else {
			logrus.Errorf("send heart beat error:%+v", err)
			reties = reties + 1
			time.Sleep(time.Duration(util.Min(reties, maxDelay)) * time.Second)
		}
	}
}
