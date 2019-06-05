package loader

import (
	v1 "github.com/celeskyking/go-nacos/api/cs/v1"
	"github.com/celeskyking/go-nacos/config"
	"github.com/celeskyking/go-nacos/types"
)

type RemoteLoader struct {
	Client v1.ConfigHttpClient
}

func NewRemoteLoader(client v1.ConfigHttpClient) Loader {
	return &RemoteLoader{
		Client: client,
	}
}

//Load 加载远程的文件信息
func (r *RemoteLoader) Load(desc *config.FileDesc) (b []byte, err error) {
	var bs []byte
	r.Client.GetConfigs(&types.ConfigsRequest{
		DataID: desc.Name,
		Tenant: desc.Namespace,
		Group:  desc.AppName + ":" + desc.Env,
	}, func(response *types.ConfigsResponse, er error) {
		if er != nil {
			err = er
			return
		}
		bs = []byte(response.Value)
		b = bs
	})
	return
}
