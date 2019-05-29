package query

import (
	"fmt"
	"testing"
)

func TestMarshal(t *testing.T) {
	u1 := &User{
		Name:"tian",
		Age:30,
		Good:true,
	}
	u2 := &MyUser{
		User:u1,
		Origin:"china",
		Group:"mfw",
	}
	result := "name=tian&age=30&good=true&group=mfw&origin=china"
	r ,err := Marshal(u2)
	fmt.Println("line:"+r)
	if err != nil {
		t.Errorf("error:%+v",err)
		t.Fail()
	}
	if result != r {
		t.Fail()
	}
}


type User struct {

	Name string `query:"name"`

	Age int `query:"age"`

	Good bool `query:"good"`
}


type MyUser struct {
	*User

	Group string `query:"group"`

	Origin string `query:"origin"`
}