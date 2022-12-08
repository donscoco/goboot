package lfucache

import (
	"container/list"
	"fmt"
	"log"
)

type LFUCache struct {
	elementMap map[string]*list.Element //用来快速定位 key-value
	//entryList *list.List
	frequency map[int]*list.List // 将相同频次的entry放一起，用来快速定位各个频次的集合

	capacity int
	min      int
}

type Entry struct {
	key       string
	val       interface{}
	frequency int
}

func New(capacity int) (l *LFUCache) {
	if capacity < 1 {
		err := fmt.Errorf("invalid capacity")
		// todo
		log.Fatal(err)
	}

	l = &LFUCache{
		elementMap: make(map[string]*list.Element),
		frequency:  make(map[int]*list.List),
		capacity:   capacity,
		min:        0,
	}
	return l
}

func (l *LFUCache) Get(key string) (val interface{}, ok bool) {
	// todo check
	if l.capacity == 0 {
		return
	}

	elem, ok := l.elementMap[key]
	if !ok {
		return nil, ok
	}

	// 移出旧的频次队列
	entry := elem.Value.(*Entry)
	oldList := l.frequency[entry.frequency] //elementMap中能拿到这里肯定能拿到。不用判空
	oldList.Remove(elem)

	entry.frequency = entry.frequency + 1

	// 放到新的频次队列
	_, ok = l.frequency[entry.frequency]
	if !ok {
		l.frequency[entry.frequency] = list.New()
	}
	newList := l.frequency[entry.frequency]
	l.elementMap[entry.key] = newList.PushFront(entry)

	// 更新最小频次
	if l.frequency[l.min].Len() == 0 && l.min == entry.frequency-1 {
		l.min = entry.frequency
	}

	return entry.val, true

}
func (l *LFUCache) Set(key string, val interface{}) {
	// todo check
	if l.capacity == 0 {
		return
	}

	elem, ok := l.elementMap[key]
	if ok {

		// 移出队列，频次+1，加入新队列
		entry := elem.Value.(*Entry)
		oldList := l.frequency[entry.frequency] // 能在 elementMap中找到，这里就一定能找到
		//if oldList == nil {
		oldList.Remove(elem)

		entry.val = val
		entry.frequency = entry.frequency + 1

		_, ok := l.frequency[entry.frequency] // 新的 频次不一定初始化了。要判断下
		if !ok {
			l.frequency[entry.frequency] = list.New()
		}
		newList := l.frequency[entry.frequency]
		//newList.PushFront(entry)
		l.elementMap[entry.key] = newList.PushFront(entry)

		if l.frequency[l.min].Len() == 0 && l.min == entry.frequency-1 {
			l.min = entry.frequency
		}

		return
	} else {

		if l.capacity == len(l.elementMap) { // 满了就清除一个最小的频次
			minList := l.frequency[l.min]
			rmElem := minList.Back() // 我们插入是front，为了公平性，从back开始移出
			minList.Remove(rmElem)
			delete(l.elementMap, rmElem.Value.(*Entry).key)
		}

		entry := &Entry{
			key:       key,
			val:       val,
			frequency: 1,
		}

		_, ok := l.frequency[entry.frequency]
		if !ok {
			l.frequency[entry.frequency] = list.New()
		}
		newList := l.frequency[entry.frequency]
		//newList.PushFront(entry)
		l.elementMap[entry.key] = newList.PushFront(entry)

		l.min = 1 // 新节点加入频次肯定是1
	}
}
func (l *LFUCache) Del(key string) {
	// todo check
	if l.capacity == 0 {
		return
	}

	elem, ok := l.elementMap[key]
	if !ok {
		return
	}
	entry := elem.Value.(*Entry)
	oldList := l.frequency[entry.frequency] // 能在 elementMap中找到，这里就一定能找到
	//if oldList == nil {
	oldList.Remove(elem)
	delete(l.elementMap, entry.key)

	// 更新最小频次,最小的已经remove了。找到第二小的 O(n)
	if l.frequency[l.min].Len() == 0 {
		l.min = 0
		for _, e := range l.elementMap {
			entry := e.Value.(*Entry)
			if entry.frequency < l.min || l.min == 0 {
				l.min = entry.frequency
			}
		}
	}

}

//func (l *LFUCache) ForTest() {
//	for _, elem := range l.elementMap {
//		entry := elem.Value.(*Entry)
//		fmt.Printf("k:%s,v:%+v,f:%d\n", entry.key, entry.val, entry.frequency)
//	}
//}
