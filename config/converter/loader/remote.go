package loader

import (
	v1 "github.com/celeskyking/go-nacos/api/cs/v1"
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
func (r *RemoteLoader) Load(desc *types.FileDesc) (b []byte, err error) {
	var bs []byte
	resp, er := r.Client.GetConfigs(&types.ConfigsRequest{
		DataID: desc.Name,
		Tenant: desc.Namespace,
		Group:  desc.AppName + ":" + desc.Env,
	})
	if er != nil {
		return nil, er
	}
	bs = []byte(resp.Value)
	b = bs
	return
}
