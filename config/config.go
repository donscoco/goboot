package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"goboot/log/mlog"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

var logger = mlog.NewLogger("config")
var ConfigFilePath string

// 节点
type node struct {
	i   int
	s   string
	f32 float32
	f64 float64
	b   bool
	o   interface{}
}

// 配置
type Config struct {
	path      string
	nodes     map[string]*node
	timestamp int64 // 更新时间
}

func NewConfiguration(path string) (c *Config) {
	c = new(Config)
	err := c.parse(path)
	if err != nil {
		logger.Errorf("解析配置错误: %s", err)
		os.Exit(1)
	}
	return
}

// 解析
func (c *Config) parse(configFile string) (err error) {
	// 读取配置文件
	filePath, err := filepath.Abs(configFile)
	if err != nil {
		//todo
		return
	}
	c.path = filePath
	jsonBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		// todo
		return
	}

	// 预处理：替换配置文件中的环境变量
	jsonString := replaceEnv(string(jsonBytes))
	// 解析到对象文件
	var configMap = make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonString), &configMap)
	if err != nil {
		//todo
		return
	}

	// 构建map
	var configNodes = make(map[string]*node, 16)
	convertNodes(configNodes, "", configMap)
	c.nodes = configNodes

	c.timestamp = time.Now().Unix()
	return nil
}

func convertNodes(tree map[string]*node, path string, obj interface{}) {
	if path == "/testI" {
		fmt.Println("")
	}
	v := reflect.ValueOf(obj)
	n := &node{}
	n.o = obj
	tree[path] = n
	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			convertNodes(tree, path+"/"+key.String(), v.MapIndex(key).Interface())
		}
	case reflect.Slice:
		tree[path+"/"+"length"] = &node{i: v.Len()}
		for i := 0; i < v.Len(); i++ {
			convertNodes(tree, path+"/"+strconv.Itoa(i), v.Index(i).Interface())
		}
	case reflect.String:
		n.s = v.Interface().(string)
	case reflect.Int64, reflect.Int32, reflect.Int, reflect.Int8, reflect.Int16:
		n.i = v.Interface().(int)
	case reflect.Float64:
		n.f64 = v.Interface().(float64)
	case reflect.Float32:
		n.f32 = v.Interface().(float32)
	case reflect.Bool:
		n.b = v.Interface().(bool)
	}
	//tree[path] = n
}

func (c *Config) GetString(path string) string {
	v, ok := c.nodes[path]
	if !ok {
		return ""
	}
	return v.s
}
func (c *Config) GetInt(path string) int {
	v, ok := c.nodes[path]
	if !ok {
		return 0
	}
	if v.i == 0 { // fixme json 转化 数字都用float64，这里的int在convert中被放到f64了
		return int(v.f64)
	}
	return v.i
}
func (c *Config) GetFloat64(path string) float64 {
	v, ok := c.nodes[path]
	if !ok {
		return 0.0
	}
	return v.f64
}
func (c *Config) GetFloat32(path string) float32 {
	v, ok := c.nodes[path]
	if !ok {
		return 0.0
	}
	return v.f32
}
func (c *Config) GetBool(path string) bool {
	v, ok := c.nodes[path]
	if !ok {
		return false
	}
	return v.b
}
func (c *Config) GetByScan(path string, objptr interface{}) (err error) {
	v, ok := c.nodes[path]
	if !ok {
		return errors.New("empty node")
	}
	data, err := json.Marshal(v.o)
	if err != nil {
		return
	}
	return json.Unmarshal(data, objptr)

}
func (c *Config) Exist(path string) bool {
	_, ok := c.nodes[path]
	if !ok {
		return false
	}
	return true
}

// replace ${var_name} macro string
var regx = regexp.MustCompile(`\${[A-Za-z0-9\-_]+}`)

func replaceEnv(origin string) string {
	return regx.ReplaceAllStringFunc(origin, func(match string) string {
		if match[2:6] == `ENV_` { // match : ${ENV_XXXXXX}
			match = os.Getenv(match[6 : len(match)-1])
		}
		return match
	})
}
