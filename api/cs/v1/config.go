package v1

import (
	"errors"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"go-nacos/api/cs"
	"go-nacos/client/http"
	"go-nacos/err"
	"go-nacos/loadbalancer"
	"go-nacos/pkg/query"
	"go-nacos/pkg/util"
	"go-nacos/types"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)




const(
	Prefix = "nacos"
	DefaultConnectTimeout = 5 * time.Second
	DefaultPollingTimeout = "60000"
	GetConfigPath = "/cs/configs"
	ListenerConfigPath = "/cs/configs/listener"
	PublishConfigPath = "/cs/configs"
	DeleteConfigPath = "/cs/configs"
	HealthPath = "/v1/cs/health"
)



type StatusCodeConverter interface {
	//
	Converter(code int) error

}

type statusCodeConverter func(code int) error


func(s statusCodeConverter) Converter(code int) error {
	return s(code)
}


func newConverter() statusCodeConverter{
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




func DefaultOption() *cs.HttpConfigOption {
	return &cs.HttpConfigOption{
		Version: "v1",
		ConnectTimeout:DefaultConnectTimeout,
	}
}




//httpClient的端口,主要用于实现http请求
type ConfigHttpClient interface {


	GetConfigs(request *types.ConfigsRequest,callback func(response *types.ConfigsResponse, err error) )


	ListenConfigs(request *types.ListenConfigsRequest, callback func(result []*types.ListenChange, err error))


	PublishConfig(request *types.PublishConfig, callback func(result *types.PublishResult, err error))


	DeleteConfigs(request *types.ConfigsRequest, callback func(response *types.DeleteResult, err error))

}


func NewConfigHttpClient(option *cs.HttpConfigOption) ConfigHttpClient {
	return newConfigHttpClient(option)
}


type configHttpClient struct {

	LB loadbalancer.LB

	Option *cs.HttpConfigOption

	Converter StatusCodeConverter
}

func newConfigHttpClient(option *cs.HttpConfigOption)  *configHttpClient {
	ss  := option.Servers
	if len(ss) == 0 {
		logrus.Errorf("不合法的nacos服务器列表,服务器最少存在一个")
		os.Exit(1)
	}
	var servers []*loadbalancer.Server
	for _, s := range ss {
		u, er := toURL(s)
		if er!= nil {
			logrus.Errorf("不合法的server地址:%s",s)
			os.Exit(1)
		}
		server := loadbalancer.NewServer(u,100, path.Join(Prefix,HealthPath))
		servers = append(servers, server)
	}
	ch := &configHttpClient{}
	if option.LBStrategy == cs.RoundRobin {
		ch.LB = loadbalancer.NewRoundRobin(servers)
	}else{
		ch.LB = loadbalancer.NewDirectProxy(servers)
	}
	ch.Option = option
	ch.Converter = NewConverter()
	return ch
}


func toURL(addr string) (*url.URL, error) {
	if !strings.HasPrefix(addr, "http://"){
		addr = "http://" + addr
	}
	u , er := url.Parse(addr)
	if er != nil {
		return nil, er
	}
	return u, nil
}



func(c *configHttpClient) GetConfigs(request *types.ConfigsRequest,callback func(response *types.ConfigsResponse, err error) ){
	logrus.Infof("get configs,request%+v",request)
	u := toUrl(c.LB)
	req, er := query.Marshal(request)
	if er != nil {
		callback(nil, er)
	}
	response, bs, errs := http.New().Timeout(DefaultConnectTimeout).Get(u+path.Join(Prefix, c.Option.Version,GetConfigPath)).Query(req).EndBytes()
	er = handleErrorResponse(c.Converter,response, errs)
	if er == nil {
		v := &types.ConfigsResponse{
			Value:string(bs),
		}
		callback(v, nil)
		return
	}else{
		callback(nil, er)
	}
}

//ListenConfigs 监听变更并且回调变更,当前的callback方法并不是纯异步的操作,只是同步操作
func(c *configHttpClient) ListenConfigs(request *types.ListenConfigsRequest, callback func(result []*types.ListenChange, err error)){
	logrus.Infof("listen configs, request:%+s", util.ToJSONString(request))
	u := toUrl(c.LB) + path.Join(Prefix, c.Option.Version, ListenerConfigPath)
	req := request.Line()
	resp, body, errs := http.New().Timeout(time.Minute).Post(u).
		Set("Long-Pulling-Timeout",DefaultPollingTimeout).
		Send("Listening-Configs="+req).
		End()
	er := handleErrorResponse(c.Converter,resp, errs)
	if er == nil {
		if len(body) == 0 {
			callback(nil, nil)
			return
		}
		lines, er := url.QueryUnescape(strings.TrimSpace(body))
		if er != nil {
			callback(nil, er)
			return
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
				callback(nil, er)
				return
			}
			c.GetConfigs(key.ToConfigsRequest(), func(response *types.ConfigsResponse, err error) {
				if err != nil {
					callback(nil, err)
					return
				}
				change.Key = key
				change.NewValue = response.Value
				changes = append(changes, change)
			})
		}
		callback(changes, nil)
	}else{
		callback(nil, er)
	}
}

//PublishConfig 发布配置信息
func(c *configHttpClient) PublishConfig(request *types.PublishConfig, callback func(result *types.PublishResult, err error)){
	logrus.Infof("publish configs, request:%+v", request)
	u := toUrl(c.LB)
	req, er := query.Marshal(request)
	if er != nil {
		callback(nil, er)
		return
	}
	u =  u + path.Join(Prefix,c.Option.Version,PublishConfigPath)
	http.New().Post(u).SendString(req).EndBytes(func(response gorequest.Response, body []byte, errs []error) {
		er = handleErrorResponse(c.Converter,response, errs)
		if er == nil {
			r, er := strconv.ParseBool(string(body))
			if er != nil {
				callback(nil, er)
				return
			}
			callback(&types.PublishResult{Success:r}, nil)
			return
		}
		callback(nil, er)
	})
}

func(c *configHttpClient) DeleteConfigs(request *types.ConfigsRequest, callback func(response *types.DeleteResult, err error)) {
	logrus.Infof("delete configs, request:%+v", request)
	u := toUrl(c.LB)
	req, er := query.Marshal(request)
	if er != nil {
		callback(nil, er)
		return
	}
	resp, bs, errs := http.New().Delete(u+path.Join(Prefix,c.Option.Version,DeleteConfigPath)).SendString(req).EndBytes()
	er = handleErrorResponse(c.Converter,resp, errs)
	if er == nil {
		r, er := strconv.ParseBool(string(bs))
		if er != nil {
			callback(nil, er)
			return
		}
		callback(&types.DeleteResult{Success:r}, nil)
	}else{
		callback(nil, er)

	}
}


func toUrl(lb loadbalancer.LB) string {
	u := lb.SelectOne().URL.String()
	if !strings.HasSuffix(u,"/"){
		u = u + "/"
	}
	return u
}


func handleErrorResponse(converter StatusCodeConverter, resp gorequest.Response, errs []error ) error {
	if resp != nil && errs != nil {
		data, er := ioutil.ReadAll(resp.Body)
		if er != nil {
			return errors.New("go-nacos system error,statusCode:5xx")
		}
		return NewHttpClientError(string(data), errs...)
	}
	if resp == nil {
		return NewHttpClientError("valid response")
	}
	return converter.Converter(resp.StatusCode)
}


type HttpClientError struct {

	Errors []error

	Message string
}

func NewHttpClientError(message string, errors ...error) *HttpClientError{
	return &HttpClientError{
		Errors: errors,
		Message: message,
	}
}

func (h *HttpClientError) Error() string{
	return fmt.Sprintf("http client apply failed, message:%s, errors:%+v",h.Message, h.Errors)
}