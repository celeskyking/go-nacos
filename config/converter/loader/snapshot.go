package loader

import (
	"gitlab.mfwdev.com/portal/go-nacos/types"
)

type SnapshotWriter interface {

	//Write 向缓存文件中写入缓存
	Write(desc *types.FileDesc, content []byte) error
}
