package naming

import (
	"gitlab.mfwdev.com/portal/go-nacos/types"
	"math/rand"
	"reflect"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//负载均衡
type InternalLB interface {

	//选择一个ServiceInstance
	SelectOne(filter func(instance *types.ServiceInstance) bool) *types.ServiceInstance

	//刷新服务器列表
	Refresh(instances []*types.ServiceInstance)

	GetAll() []*types.ServiceInstance
}

//NewRandom 返回一个随机选择的负载均衡算法,后续扩展更多算法
func NewRandom() InternalLB {
	return &Random{}
}

type Random struct {
	Servers []*types.ServiceInstance
}

func (r *Random) Refresh(instances []*types.ServiceInstance) {
	if len(r.Servers) == 0 {
		r.Servers = instances
	}
	if !reflect.DeepEqual(r.Servers, instances) {
		r.Servers = instances
	}
}

func (r *Random) GetAll() []*types.ServiceInstance {
	return r.Servers
}

func (r *Random) SelectOne(filter func(instance *types.ServiceInstance) bool) *types.ServiceInstance {
	var servers []*types.ServiceInstance
	for _, i := range r.Servers {
		if filter == nil || filter(i) {
			servers = append(servers, i)
		}
	}
	l := len(servers)
	if l == 0 || servers == nil {
		return nil
	}
	return servers[rand.Intn(l-1)]
}
