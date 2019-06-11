package api

import (
	"gitlab.mfwdev.com/portal/go-nacos/client/loadbalancer"
	"net/url"
	"strings"
)

func ToURL(addr string) (*url.URL, error) {
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}
	u, er := url.Parse(addr)
	if er != nil {
		return nil, er
	}
	return u, nil
}

func SelectOne(lb loadbalancer.LB) string {
	u := lb.SelectOne().URL.String()
	if !strings.HasSuffix(u, "/") {
		u = u + "/"
	}
	return u
}
