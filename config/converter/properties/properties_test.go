package properties

import (
	"fmt"
	"gitlab.mfwdev.com/portal/go-nacos/config"
	"gitlab.mfwdev.com/portal/go-nacos/types"
	"testing"
)

func TestMapFile_Desc(t *testing.T) {
	c := config.GetConverter("properties")
	mapfile := c.Convert(&types.FileDesc{
		Name:      "demo.properties",
		AppName:   "app1",
		Env:       "beta",
		Namespace: "demo",
	}, []byte("name=tianqing.wang")).(*MapFile)
	fmt.Println(mapfile.MD5())
	fmt.Println(mapfile.MustGet("name"))
}
