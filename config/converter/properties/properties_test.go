package properties

import (
	"fmt"
	"github.com/celeskyking/go-nacos/config"
	"testing"
)

func TestMapFile_Desc(t *testing.T) {
	c := config.GetConverter("properties")
	mapfile := c.Convert(&config.FileDesc{
		Name:      "demo.properties",
		AppName:   "app1",
		Env:       "beta",
		Namespace: "demo",
	}, []byte("name=tianqing.wang")).(*MapFile)
	fmt.Println(mapfile.MD5())
	fmt.Println(mapfile.MustGet("name"))
}
