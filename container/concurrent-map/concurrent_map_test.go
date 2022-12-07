package concurrent_map

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {

	ctx, cancel := context.WithCancel(context.TODO())
	var wg sync.WaitGroup
	cm := New(32, nil)
	tm := NewTM()

	threadNum := 8

	//cm.Set("1", 1)
	//fmt.Println(cm.Get("1"))

	for i := 0; i < threadNum; i++ {
		cm.Set("key"+strconv.Itoa(i), 0)
		tm.set("key"+strconv.Itoa(i), 0)
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					v, ok := cm.Get(key)
					if !ok { // 第一次不存在先给他加上。
						cm.Set(key, 0)
					}
					cm.Set(key, v.(int)+1)
				}

			}
		}("key" + strconv.Itoa(i))

		go func(key string) {
			for {
				select {
				case <-ctx.Done():
				default:
					tm.set(key, tm.get(key).(int)+1)
				}
			}
		}("key" + strconv.Itoa(i))
	}

	time.Sleep(1 * time.Second)
	cancel()
	wg.Wait()

	count1 := 0
	count2 := 0
	for i := 0; i < threadNum; i++ {
		v, _ := cm.Get("key" + strconv.Itoa(i))
		count1 = count1 + v.(int)

		v2 := tm.get("key" + strconv.Itoa(i))
		count2 = count2 + v2.(int)
	}
	println("conc map write count:", count1)
	println("lock map write count:", count2)
}

type TestMap struct {
	data map[string]interface{}
	sync.Mutex
}

func NewTM() (tm *TestMap) {
	return &TestMap{
		data: make(map[string]interface{}),
	}
}
func (m *TestMap) get(key string) interface{} {
	m.Lock()
	defer m.Unlock()

	return m.data[key]
}
func (m *TestMap) set(key string, val interface{}) {
	m.Lock()
	defer m.Unlock()

	m.data[key] = val
}
func (m *TestMap) del(key string) {
	m.Lock()
	defer m.Unlock()

	delete(m.data, key)

}
