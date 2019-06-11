package loader

import (
	"gitlab.mfwdev.com/portal/go-nacos/types"
)

type Loader interface {

	//加载文件
	Load(desc *types.FileDesc) ([]byte, error)
}
