package loader

import "github.com/celeskyking/go-nacos/config"

type Loader interface {

	//加载文件
	Load(desc *config.FileDesc) ([]byte, error)
}
