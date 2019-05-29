package app

import (
	"github.com/celeskyking/go-nacos/config/service"
)

type Config struct {
	//app名称
	AppName string
	//当前的环境
	Env string
	//当前的namespace
	Namespace string
}

//健康监测
type HealthCheck interface {
}

type Application struct {
	//配置信息
	Config *Config
}

func (a *Application) Start() {

}

func (a *Application) ConfigClient() *service.ConfigService {
	return nil
}
