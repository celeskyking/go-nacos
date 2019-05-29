package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

func MD5(b []byte) string{
	if len(b)==0 {
		return ""
	}
	m := md5.New()
	m.Write(b)
	return hex.EncodeToString(m.Sum(nil))
}

func Min(i,j int) int {
	if i<j {
		return i
	}
	return j
}


func ToJSONString(object interface{}) string {
	data, _ := json.Marshal(object)
	return string(data)
}