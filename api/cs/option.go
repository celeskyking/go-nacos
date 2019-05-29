package cs

import "time"

type HttpConfigOption struct {
	//连接超时
	ConnectTimeout time.Duration
	//当前nacos的api版本
	Version string
	//服务列表
	Servers []string

	LBStrategy LBStrategy

}


type LBStrategy int

const(

	Direct LBStrategy = iota

	RoundRobin
)