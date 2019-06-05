package err

import (
	"fmt"
	"github.com/pkg/errors"
)

var ErrBadRequest = errors.New("客户端请求中的语法错误")

var ErrForbidden = errors.New("没有权限")

var ErrNotFound = errors.New("无法找到配置文件资源")

var ErrInternalServerError = errors.New("服务器内部错误")

var ErrUnSupportStatusCode = errors.New("不支持的statusCode")

var ErrKeyNotValid = errors.New("配置变更的key不合法")

var ErrNotPropertiesFile = errors.New("解析properties文件失败")

var ErrKeyNotFound = errors.New("key not found")

var ErrFileNotFound = errors.New("file not found")

var ErrLoaderNotWork = errors.New("loader not work")

var ErrNamingService = errors.New("register service error")

type HttpClientError struct {
	Errors []error

	Message string
}

func NewHttpClientError(message string, errors ...error) *HttpClientError {
	return &HttpClientError{
		Errors:  errors,
		Message: message,
	}
}

func (h *HttpClientError) Error() string {
	return fmt.Sprintf("http client apply failed, message:%s, errors:%+v", h.Message, h.Errors)
}
