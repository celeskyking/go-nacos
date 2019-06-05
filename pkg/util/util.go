package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net"
	"os"
	"strings"
)

func MD5(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	m := md5.New()
	m.Write(b)
	return hex.EncodeToString(m.Sum(nil))
}

func Min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func ToJSONString(object interface{}) string {
	data, _ := json.Marshal(object)
	return string(data)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func LocalIP() string {
	localIP := os.Getenv("LOCAL_IP")
	if localIP == "" {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return ""
		}
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					localIP = ipnet.IP.String()
					break
				}
			}
		}
	}
	return localIP
}

func ToMap(text string) map[string]string {
	if text == "" {
		return nil
	}
	parts := strings.Split(text, ",")
	m := make(map[string]string)
	if len(parts) == 0 {
		return m
	}
	for _, p := range parts {
		keyAndValue := strings.Split(p, "=")
		if len(keyAndValue) != 2 {
			return nil
		}
		m[keyAndValue[0]] = keyAndValue[1]
	}
	return m
}

func MapToString(params map[string]string) string {
	var parts []string
	for key, value := range params {
		parts = append(parts, key+"="+value)
	}
	return strings.Join(parts, ",")
}
