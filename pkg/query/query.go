package query

import (
	"fmt"
	"reflect"
	"strings"
)

func Marshal(object interface{})( string,error){
	var result []string
	r, err := marshal(object,result)
	if err != nil {
		return "", err
	}
	return strings.Join(r,"&"), nil
}


func marshal(object interface{}, result []string) ([]string, error) {
	t := reflect.TypeOf(object)
	v := reflect.ValueOf(object).Elem()
	// 取指针指向的结构体变量
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i:=0; i< t.NumField(); i ++ {
		f := t.Field(i)
		q := f.Tag.Get("query")
		vv := v.Field(i)
		if vv.Kind() == reflect.Ptr {
			s, err := marshal(vv.Interface(),result)
			if err != nil {
				return nil, err
			}
			result = append(result, s...)
			continue
		}
		if q == "" {
			result = append(result, f.Name+"="+fmt.Sprintf("%v",v.Field(i).Interface()))
		}else{
			result = append(result, q+"="+fmt.Sprintf("%v",v.Field(i).Interface()))
		}
	}
	return result, nil
}

