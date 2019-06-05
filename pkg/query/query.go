package query

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

var v *validator.Validate

type Transfer interface {
	Transfer(value reflect.Value) string
}

func init() {
	v = validator.New()
}

type TransferFunc func(value reflect.Value) string

func (tf TransferFunc) Transfer(value reflect.Value) string {
	return tf(value)
}

var transfers = map[string]TransferFunc{
	"json": func(value reflect.Value) string {
		data, _ := json.Marshal(value.Interface())
		return string(data)
	},
	"simple": func(value reflect.Value) string {
		return fmt.Sprintf("%v", value.Interface())
	},
}

func GetString(transfer string, value reflect.Value) string {
	return transfers[transfer].Transfer(value)
}

func Marshal(object interface{}) (string, error) {
	er := v.Struct(object)
	if er != nil {
		return "", errors.Wrap(er, "query validator")
	}
	var result []string
	r, err := marshal(object, result)
	if err != nil {
		return "", err
	}
	return strings.Join(r, "&"), nil
}

func marshal(object interface{}, result []string) ([]string, error) {
	t := reflect.TypeOf(object)
	v := reflect.ValueOf(object).Elem()
	// 取指针指向的结构体变量
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		q := f.Tag.Get("query")
		vv := v.Field(i)
		if vv.Kind() == reflect.Ptr && q == "inline" {
			s, err := marshal(vv.Interface(), result)
			if err != nil {
				return nil, err
			}
			result = append(result, s...)
			continue
		}
		transfer := f.Tag.Get("transfer")
		if transfer == "" {
			transfer = "simple"
		}
		if q == "" {
			result = append(result, f.Name+"="+GetString(transfer, v.Field(i)))
		} else {
			result = append(result, q+"="+GetString(transfer, v.Field(i)))
		}
	}
	return result, nil
}
