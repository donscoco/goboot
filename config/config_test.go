package config

import (
	"fmt"
	"testing"
)

// go test -v -run TestConfig ./config
func TestConfig(t *testing.T) {
	fmt.Println("test")
	config := NewConfiguration("./config-demo.json")
	fmt.Println(config.Exist("/testI"))
	fmt.Println(config.GetInt("/testI"))
	fmt.Println(config.GetFloat64("/testF"))
	fmt.Println(config.GetFloat32("/testF"))
	fmt.Println(config.GetBool("/testB"))
	fmt.Println(config.GetString("/testO/Name"))

	obj := make(map[string]interface{})
	config.GetByScan("/testO", &obj)
	fmt.Println(obj)
}
