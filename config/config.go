package config

import (
	"sync"
)

type FileConverters map[string]FileConverter

var F FileConverters

var once sync.Once

func (fc *FileConverters) Register(fileType string, converter FileConverter) {
	F[fileType] = converter
}

func NewFileConverters() FileConverters {
	return make(map[string]FileConverter, 0)
}

func init() {
	once.Do(func() {
		F = NewFileConverters()
	})
}

//RegisterConverter 注册转化器,暂时不支持卸载的逻辑
func RegisterConverter(fileType string, converter FileConverterFunc) {
	F.Register(fileType, converter)
}

func GetConverter(name string) FileConverter {
	if v, ok := F[name]; ok {
		return v
	}
	return nil
}

type FileConverter interface {
	//转换器
	Convert(desc *FileDesc, content []byte) FileMirror
}

type FileConverterFunc func(desc *FileDesc, content []byte) FileMirror

func (f FileConverterFunc) Convert(desc *FileDesc, content []byte) FileMirror {
	return f(desc, content)
}
