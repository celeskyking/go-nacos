package loader

import "github.com/celeskyking/go-nacos/config"

type SnapshotWriter interface {

	//Write 向缓存文件中写入缓存
	Write(desc *config.FileDesc, content []byte) error
}
