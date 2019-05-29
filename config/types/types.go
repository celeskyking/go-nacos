package types

type FileMirror interface {

	//写入所有的数据
	OnChanged(notifyC <- chan []byte)

	//获取所有的文件内容,不关心格式
	GetContent() []byte

	//文件描述
	Desc() *FileDesc

	//文件的md5值
	MD5() string

}

type FileDesc struct {
	//
	Name string
	//应用名
	AppName string
	//环境
	Env string

	Namespace string

}