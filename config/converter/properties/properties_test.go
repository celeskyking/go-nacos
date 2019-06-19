package properties

import (
	"fmt"
	"github.com/celeskyking/go-nacos/config/converter"
	"github.com/celeskyking/go-nacos/types"
	"testing"
)

func TestMapFile_Desc(t *testing.T) {
	c := converter.GetConverter("properties")
	mapfile := c.Convert(&types.FileDesc{
		Name:      "demo.properties",
		Group:     "app1",
		Namespace: "demo",
	}, []byte("name=tianqing.wang")).(*MapFile)
	fmt.Println(mapfile.MD5())
	fmt.Println(mapfile.MustGet("name"))
}
