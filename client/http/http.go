package http

import (
	"github.com/parnurzeal/gorequest"
	"sync"
)

var Client *gorequest.SuperAgent

var once sync.Once

func init() {
	once.Do(func() {
		Client = gorequest.New()
	})
}


func New() *gorequest.SuperAgent{
	return Client.Clone()
}
