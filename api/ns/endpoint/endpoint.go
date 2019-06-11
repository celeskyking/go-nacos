package endpoint

import (
	"bufio"
	"bytes"
	"github.com/sirupsen/logrus"
	"gitlab.mfwdev.com/portal/go-nacos/client/http"
	"path"
	"time"
)

const (
	ServerListPath string = "nacos/serverlist"
	Interval              = 30 * time.Second
)

func NewEndpoint(address string) *Endpoint {
	return &Endpoint{
		address: address,
	}
}

type Endpoint struct {
	//endpoint的地址
	address string
}

func (e *Endpoint) Run(stop <-chan struct{}) chan []string {
	notify := make(chan []string, 0)
	ticker := time.NewTicker(Interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				resp, data, errs := http.New().Get("http://" + path.Join(e.address, ServerListPath)).EndBytes()
				if len(errs) != 0 {
					for _, e := range errs {
						logrus.Errorf("endpoint failed:%+v", e)
						return
					}
				}
				if resp.StatusCode != 200 {
					logrus.Error("endpoint statusCode not ok")
					return
				}
				var servers []string
				s := bufio.NewScanner(bytes.NewReader(data))
				for s.Scan() {
					servers = append(servers, s.Text())
				}
				notify <- servers
			case <-stop:
				close(notify)
				return
			}
		}
	}()
	return notify
}
