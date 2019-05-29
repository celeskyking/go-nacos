package types

import (
	"go-nacos/err"
	"strings"
)

const(
	FieldSeparator  byte = 2
	ArticleSeparator  byte = 1
)


type ConfigsRequest struct {

	Tenant string `query:"tenant"`

	DataID string `query:"dataId"`

	Group string `query:"group"`

}


type ConfigsResponse struct {
	Value string
}


type ListenConfigsRequest struct {
	//监听的配置信息
	ListeningConfigs []*ListenKey

}


func(lcs *ListenConfigsRequest) Line() string {
	var result string
	for _, k := range lcs.ListeningConfigs {
		result += k.Line()
	}
	return result
}


type ListenChange struct {

	Key *ListenKey

	//新值
	NewValue string

}


type ListenKey struct {

	DataID string

	Group string

	ContentMD5 string

	Tenant string
}


func(l *ListenKey) Line() string{
	var result []string
	result = append(result, l.DataID)
	result = append(result, l.Group)
	result = append(result, l.ContentMD5)
	if l.Tenant!="" {
		result = append(result, l.Tenant)
	}
	return strings.Join(result, string(FieldSeparator))+ string(ArticleSeparator)
}


func ParseListenKey(line string) (*ListenKey, error){
	line = strings.Trim(line,string(ArticleSeparator))
	parts := strings.Split(line, string(FieldSeparator))
	l := len(parts)
	if !(l==3 || l==4) {
		return nil, err.ErrKeyNotValid
	}
	key  :=  &ListenKey{
		DataID: parts[0],
		Group: parts[1],
	}
	if l ==3 {
		key.Tenant = parts[2]
	}else{
		key.ContentMD5 = parts[2]
		key.Tenant = parts[3]
	}
	return key, nil
}


func(l *ListenKey) ToConfigsRequest() *ConfigsRequest {
	return &ConfigsRequest{
		DataID:l.DataID,
		Group:l.Group,
		Tenant:l.Tenant,
	}
}


type PublishConfig struct {

	Tenant string `query:"tenant"`

	DataID string `query:"dataId"`

	Group string `query:"group"`

	Content string `query:"content"`

}


type PublishResult struct {

	Success bool

}



type DeleteResult struct {

	Success bool
}