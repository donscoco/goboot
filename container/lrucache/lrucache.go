package lrucache

import (
	"container/list"
	"fmt"
	"log"
)

type LRUCache struct {
	elementList *list.List
	elementMap  map[string]*list.Element
	size        int
}

type Entry struct { // 用于移除的时候快速找到key
	key string
	val interface{}
}

func New(size int) *LRUCache {
	if size < 1 {
		err := fmt.Errorf("invalid lrucache size")
		// todo
		log.Fatal(err)
	}
	return &LRUCache{
		elementList: list.New(),
		elementMap:  make(map[string]*list.Element, size),
		size:        size,
	}
}
func (l *LRUCache) Set(key string, val interface{}) {
	elem, ok := l.elementMap[key]
	if ok { // 存在,移动到头部
		elem.Value.(*Entry).val = val
		l.elementList.MoveToFront(elem)
	} else { // 不存在,创建,加入,移除尾部
		entry := &Entry{
			key: key,
			val: val,
		}
		elem := l.elementList.PushFront(entry)
		l.elementMap[key] = elem

		if l.size < l.elementList.Len() {
			rmElem := l.elementList.Back()
			delete(l.elementMap, rmElem.Value.(*Entry).key)
			l.elementList.Remove(rmElem)
		}
	}
}
func (l *LRUCache) Get(key string) (val interface{}, ok bool) {
	elem, ok := l.elementMap[key]
	if !ok {
		return nil, ok
	}

	// 存在就是进行了一次访问，直接移动到队头
	l.elementList.MoveToFront(elem)
	kv := elem.Value.(*Entry)

	return kv.val, true
}

func (l *LRUCache) Len() int {
	return l.elementList.Len()
}

// 测试使用的。查看 list 的顺序
//func (l *LRUCache) ForTeset() {
//	len := l.elementList.Len()
//	for i := 0; i < len; i++ {
//		elem := l.elementList.Front()
//		log.Println(elem.Value)
//		l.elementList.Remove(elem)
//	}
//}
