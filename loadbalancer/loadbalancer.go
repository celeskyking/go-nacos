package loadbalancer

import (
	"github.com/celeskyking/go-nacos/client/http"
	"github.com/parnurzeal/gorequest"
	"net/url"
	"sync"
	"time"
)

const (
	DefaultTimeout  = 3 * time.Second
	DefaultInterval = 10 * time.Second
)

type ServerHealthState int

const (
	Passing ServerHealthState = iota
	Critical
)

//LB负载均衡接口,提供
type LB interface {
	//SelectOne 选择一个端口
	SelectOne() *Server
}

type DirectProxy struct {
	Server *Server
}

func (d *DirectProxy) SelectOne() *Server {
	return d.Server
}

func NewDirectProxy(server []*Server) LB {
	if len(server) == 0 {
		return &DirectProxy{}
	}
	return &DirectProxy{
		Server: server[0],
	}
}

type HealthCheck interface {
	//
	Check(stop <-chan struct{}, server *Server, callback func(state ServerHealthState))
}

type healthCheck struct {
}

func (h *healthCheck) Check(stop <-chan struct{}, server *Server, callback func(state ServerHealthState)) {
	ticker := time.NewTicker(DefaultInterval)
	currentState := Passing
	for {
		select {
		case <-stop:
		case <-ticker.C:
			http.New().Clone().Get(server.URL.String() + server.HealthPath).End(func(response gorequest.Response, body string, errs []error) {
				if len(errs) != 0 || response == nil || response.StatusCode != 200 {
					if currentState != Critical {
						callback(Critical)
						currentState = Critical
					}
					return
				}
				if response.StatusCode == 200 {
					if currentState == Critical {
						callback(Passing)
						currentState = Passing
					}
				}
			})
		}
	}
}

type RoundRobin struct {
	//权重最大公约数
	GcdWeight int
	//服务器列表
	Servers []*Server
	//当前的索引
	CurrentIndex int
	//当前的权重
	CurrentWeight int

	lock sync.Mutex

	Stop chan struct{}
}

// 最大公约数
func greaterCommonDivisor(i, j int) int {
	for j != 0 {
		i, j = j, i%j
	}
	return i
}

func (r *RoundRobin) SelectOne() *Server {
	r.lock.Lock()
	defer r.lock.Unlock()
	servers := r.selectHealthyServers()
	if len(servers) == 0 {
		//todo 降级server
		return nil
	}
	if len(servers) == 1 {
		return servers[0]
	}
	for {
		r.CurrentIndex = (r.CurrentIndex + 1) % len(servers)
		if r.CurrentIndex == 0 {
			r.CurrentWeight = r.CurrentWeight - r.GcdWeight
			if r.CurrentWeight <= 0 {
				r.CurrentWeight = r.getMaxWeightForServers()
				if r.CurrentWeight == 0 {
					return nil
				}
			}
		}
		if weight := servers[r.CurrentIndex].Weight; weight > r.CurrentWeight {
			return servers[r.CurrentIndex]
		}
	}
}

func (r *RoundRobin) selectHealthyServers() []*Server {
	var result []*Server
	for _, s := range r.Servers {
		if s.State == Passing {
			result = append(result, s)
		}
	}
	return result
}

func (r *RoundRobin) getMaxWeightForServers() int {
	servers := r.selectHealthyServers()
	max := 0
	for i := 0; i < len(servers); i++ {
		if weight := servers[i].Weight; weight > max {
			max = weight
		}
	}
	return max
}

func (r *RoundRobin) getGcdForServers() int {
	m := 0
	servers := r.selectHealthyServers()
	for i := 0; i < len(servers); i++ {
		if i == 0 {
			m = servers[i].Weight
		} else {
			m = greaterCommonDivisor(servers[i].Weight, servers[i+1].Weight)
		}
	}
	return m
}

//NewRoundRobin 加权轮询服务器
//todo 支持动态刷新servers,目前不支持
func NewRoundRobin(servers []*Server) LB {
	r := &RoundRobin{
		Servers: servers,
		Stop:    make(chan struct{}, 0),
	}
	if len(servers) > 0 {
		r.Start(r.Stop)
	}
	r.refresh()
	return r
}

func (r *RoundRobin) Start(stop <-chan struct{}) {
	for _, s := range r.Servers {
		go func(stop <-chan struct{}) {
			checker := &healthCheck{}
			checker.Check(stop, s, func(state ServerHealthState) {
				s.State = state
				r.refresh()
			})
		}(stop)
	}
}

func (r *RoundRobin) refresh() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.GcdWeight = r.getGcdForServers()
	r.CurrentWeight = r.getMaxWeightForServers()
	r.CurrentIndex = -1
}

type Server struct {
	//地址
	URL *url.URL
	//权重
	Weight int

	State ServerHealthState

	//健康监测的地址
	HealthPath string
}

func NewServer(url *url.URL, weight int, healthPath string) *Server {
	return &Server{
		URL:        url,
		Weight:     weight,
		State:      Passing,
		HealthPath: healthPath,
	}
}
