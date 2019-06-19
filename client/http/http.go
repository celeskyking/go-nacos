package http

import (
	"github.com/parnurzeal/gorequest"
	"github.com/satori/go.uuid"
	"sync"
)

var Client *gorequest.SuperAgent

var once sync.Once

func init() {
	once.Do(func() {
		Client = gorequest.New()
	})
}

func NewNamingHttp() *gorequest.SuperAgent {
	return Client.Clone().Set("User-Agent", "nacos-go-sdk:v1.0.1").
		Set("Client-Version", "nacos-go-sdk:v1.0.1").
		Set("Connection", "Keep-Alive").
		Set("RequestId", uid()).
		Set("Request-Module", "Naming")
}

func New() *gorequest.SuperAgent {
	return Client.Clone()
}

func uid() string {
	u := uuid.NewV4()
	return u.String()
}
