package err

import "github.com/pkg/errors"

var ErrBadRequest = errors.New("客户端请求中的语法错误")

var ErrForbidden = errors.New("没有权限")

var ErrNotFound = errors.New("无法找到配置文件资源")

var ErrInternalServerError = errors.New("服务器内部错误")

var ErrUnSupportStatusCode = errors.New("不支持的statusCode")


var ErrKeyNotValid = errors.New("配置变更的key不合法")


var ErrNotPropertiesFile = errors.New("解析properties文件失败")


var ErrkeyNotFound = errors.New("key not found")