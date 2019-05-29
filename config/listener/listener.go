package listener

import (
	"go-nacos/config/types"
)

type EventType int

const(
	//新增
	Add EventType = iota
	//删除
	Delete

	Update
)



//用来监听整个文件都发生了变更
type FileListener interface {

	//当文件变更的时候触发
	OnChange( oldContent, newContent []byte, ctx *types.FileDesc)


}


//ValueListener 用来监听指定key的value发生了变更
type ValueListener interface {

	//OnChange 当发生值变更的时候会触发
	OnChange(key string, curValue, newValue string, ctx *types.FileDesc)

}


type FileListenerFunc func(oldContent, newContent []byte, ctx *types.FileDesc)


func(f FileListenerFunc) OnChange(oldContent, newContent []byte, ctx *types.FileDesc) {
	f(oldContent,newContent,ctx)
}


type ValueListenerFunc func(key string, curValue, newValue string, ctx *types.FileDesc)


func(f ValueListenerFunc) OnChange(key string, curValue, newValue string, ctx *types.FileDesc) {
	f(key, curValue,newValue,ctx)
}








