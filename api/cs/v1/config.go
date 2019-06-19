package v1

import (
	"errors"
	"fmt"
	"github.com/celeskyking/go-nacos/api"
	"github.com/celeskyking/go-nacos/client/http"
	"github.com/celeskyking/go-nacos/client/loadbalancer"
	"github.com/celeskyking/go-nacos/err"
	"github.com/celeskyking/go-nacos/pkg/query"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/celeskyking/go-nacos/types"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	Prefix                = "nacos"
	DefaultConnectTimeout = 5 * time.Second
	DefaultPollingTimeout = "60000"
	GetConfigPath         = "/cs/configs"
	ListenerConfigPath    = "/cs/configs/listener"
	PublishConfigPath     = "/cs/configs"
	DeleteConfigPath      = "/cs/configs"
	HealthPath            = "/v1/console/health/liveness"
)

type StatusCodeConverter interface {
	//
	Converter(code int) error
}

type statusCodeConverter func(code int) error

func (s statusCodeConverter) Converter(code int) error {
	return s(code)
}

func newConverter() statusCodeConverter {
	return func(code int) error {
		switch code {
		case 200:
			return nil
		case 400:
			return err.ErrBadRequest
		case 403:
			return err.ErrForbidden
		case 404:
			return err.ErrNotFound
		case 500:
			return err.ErrInternalServerError
		default:
			return err.ErrUnSupportStatusCode
		}
	}
}

func NewConverter() StatusCodeConverter {
	return newConverter()
}

//httpClient的端口,主要用于实现http请求
type ConfigHttpClient interface {
	GetConfigs(request *types.ConfigsRequest) (response *types.ConfigsResponse, err error)
	//ListenConfigs 是一个阻塞方法,最好不要在主线程内调用,异步调用比较可靠,或者使用高级API
	ListenConfigs(request *types.ListenConfigsRequest) (result []*types.ListenChange, err error)

	PublishConfig(request *types.PublishConfig) (result *types.Result, err error)

	DeleteConfigs(request *types.ConfigsRequest) (response *types.Result, err error)
}

func NewConfigHttpClient(option *api.HttpConfigOption) ConfigHttpClient {
	return newConfigHttpClient(option)
}

type configHttpClient struct {
	LB loadbalancer.LB

	Option *api.HttpConfigOption

	Converter StatusCodeConverter
}

func newConfigHttpClient(option *api.HttpConfigOption) *configHttpClient {
	ss := option.Servers
	if len(ss) == 0 {
		logrus.Errorf("不合法的nacos服务器列表,服务器最少存在一个")
		os.Exit(1)
	}
	var servers []*loadbalancer.Server
	for _, s := range ss {
		u, er := api.ToURL(s)
		if er != nil {
			logrus.Errorf("不合法的server地址:%s", s)
			os.Exit(1)
		}
		server := loadbalancer.NewServer(u, 100, path.Join(Prefix, HealthPath))
		servers = append(servers, server)
	}
	ch := &configHttpClient{}
	if option.LBStrategy == api.RoundRobin {
		ch.LB = loadbalancer.NewRoundRobin(servers, true)
	} else {
		ch.LB = loadbalancer.NewDirectProxy(servers)
	}
	ch.Option = option
	ch.Converter = NewConverter()
	return ch
}

func (c *configHttpClient) GetConfigs(request *types.ConfigsRequest) (*types.ConfigsResponse, error) {
	logrus.Infof("get configs,request%+v", request)
	u := api.SelectOne(c.LB)
	if request.Tenant == "Public" {
		request.Tenant = ""
	}
	req, er := query.Marshal(request)
	if er != nil {
		return nil, er
	}
	response, bs, errs := http.New().Timeout(DefaultConnectTimeout).Get(u + path.Join(Prefix, c.Option.Version, GetConfigPath)).Query(req).EndBytes()
	er = handleErrorResponse(c.Converter, response, errs)
	if er == nil {
		v := &types.ConfigsResponse{
			Value: string(bs),
		}
		return v, nil
	} else {
		return nil, er
	}
}

//ListenConfigs 监听变更并且回调变更,当前的callback方法并不是纯异步的操作,只是同步操作
func (c *configHttpClient) ListenConfigs(request *types.ListenConfigsRequest) ([]*types.ListenChange, error) {
	logrus.Infof("listen configs, request:%+s", util.ToJSONString(request))
	u := api.SelectOne(c.LB) + path.Join(Prefix, c.Option.Version, ListenerConfigPath)
	req := request.Line()
	fmt.Println("lines:" + req)
	resp, body, errs := http.New().Timeout(time.Minute).Post(u).
		Set("Long-Pulling-Timeout", DefaultPollingTimeout).
		Send("Listening-Configs=" + req).
		End()
	er := handleErrorResponse(c.Converter, resp, errs)
	if er == nil {
		if len(body) == 0 {
			return nil, nil
		}
		lines, er := url.QueryUnescape(strings.TrimSpace(body))
		if er != nil {
			return nil, er
		}
		parts := strings.Split(lines, string(types.ArticleSeparator))
		var changes []*types.ListenChange
		for _, p := range parts {
			if len(p) == 0 {
				continue
			}
			change := &types.ListenChange{}
			key, er := types.ParseListenKey(p)
			if er != nil {
				return nil, er
			}
			resp, er := c.GetConfigs(key.ToConfigsRequest())
			if er != nil {
				return nil, er
			}
			change.Key = key
			change.NewValue = resp.Value
			changes = append(changes, change)
		}
		return changes, nil
	} else {
		return nil, er
	}
}

//PublishConfig 发布配置信息
func (c *configHttpClient) PublishConfig(request *types.PublishConfig) (*types.Result, error) {
	logrus.Infof("publish configs, request:%+v", request)
	u := api.SelectOne(c.LB)
	req, er := query.Marshal(request)
	if er != nil {
		return nil, er
	}
	u = u + path.Join(Prefix, c.Option.Version, PublishConfigPath)
	resp, body, errs := http.New().Post(u).SendString(req).EndBytes()
	er = handleErrorResponse(c.Converter, resp, errs)
	if er == nil {
		r, er := strconv.ParseBool(string(body))
		if er != nil {
			return nil, er
		}
		return &types.Result{Success: r}, nil
	}
	return nil, er
}

func (c *configHttpClient) DeleteConfigs(request *types.ConfigsRequest) (*types.Result, error) {
	logrus.Infof("delete configs, request:%+v", request)
	u := api.SelectOne(c.LB)
	req, er := query.Marshal(request)
	if er != nil {
		return nil, er
	}
	resp, bs, errs := http.New().Delete(u + path.Join(Prefix, c.Option.Version, DeleteConfigPath)).SendString(req).EndBytes()
	er = handleErrorResponse(c.Converter, resp, errs)
	if er == nil {
		r, er := strconv.ParseBool(string(bs))
		if er != nil {
			return nil, er
		}
		return &types.Result{Success: r}, nil
	} else {
		return nil, er
	}
}

func handleErrorResponse(converter StatusCodeConverter, resp gorequest.Response, errs []error) error {
	if resp != nil && errs != nil {
		data, er := ioutil.ReadAll(resp.Body)
		if er != nil {
			return errors.New("go-nacos system error,statusCode:5xx")
		}
		return err.NewHttpClientError(string(data), errs...)
	}
	if resp == nil {
		return err.NewHttpClientError("valid response")
	}
	return converter.Converter(resp.StatusCode)
}
