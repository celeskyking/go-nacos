package loader

import (
	"github.com/celeskyking/go-nacos/types"
)

type Loader interface {

	//加载文件
	Load(desc *types.FileDesc) ([]byte, error)
}
