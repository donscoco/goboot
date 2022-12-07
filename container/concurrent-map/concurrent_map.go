package concurrent_map

import (
	"fmt"
	"hash/crc32"
	"log"
	"sync"
)

/*
sync.map 是使用读写锁去实现的，那种map适合读多写少。
这里实现一个分片的map，
核心思想是利用 TCMalloc 的核心思想，细化内存粒度，让各个线程之间不会放因为同一片内存冲突
所以本质就是把一个map给分片成多个map，key过来后hash处理各自的map。只要不在同一个分片上。线程之间就能并行的处理。
*/
type ConcurrentMap struct {
	shardings []*partition // 分片越多，并发冲突概率越低

	// hash func 负责指定key的hash形式,给出这个key应该在那个分片上,现在先简单直接%。后续考虑是否有必要使用上一致性hash，还有节点变更等问题问题。
	hashfunc func(key string, shardings []*partition) *partition
}

func New(partitionNum int, f func(key string, shardings []*partition) *partition) (cm *ConcurrentMap) {
	if partitionNum < 1 {
		// todo
		err := fmt.Errorf("invalid partitionNum")
		log.Fatal(err)
	}
	cm = &ConcurrentMap{
		shardings: make([]*partition, 0, partitionNum),
		hashfunc:  f,
	}
	for i := 0; i < partitionNum; i++ {
		partition := &partition{
			data: make(map[string]interface{}, 16),
		}
		cm.shardings = append(cm.shardings, partition)
	}
	if f == nil {
		f = defaultHash
	}
	cm.hashfunc = f
	return
}
func defaultHash(key string, shardings []*partition) *partition {
	var index uint32
	if len(key) < 64 {
		var scratch [64]byte
		copy(scratch[:], key)
		index = crc32.ChecksumIEEE(scratch[:len(key)])
	} else {
		index = crc32.ChecksumIEEE([]byte(key))
	}
	return shardings[int(index)%len(shardings)]
}

func (m *ConcurrentMap) Get(key string) (val interface{}, ok bool) {
	p := m.hashfunc(key, m.shardings)
	return p.get(key)
}
func (m *ConcurrentMap) Set(key string, val interface{}) {
	p := m.hashfunc(key, m.shardings)
	p.set(key, val)
	return
}
func (m *ConcurrentMap) Del(key string) {
	p := m.hashfunc(key, m.shardings)
	p.del(key)
}

type partition struct {
	data map[string]interface{}
	sync.Mutex
}

func (p *partition) get(key string) (val interface{}, ok bool) {
	p.Lock()
	defer p.Unlock()

	val, ok = p.data[key]
	return
}
func (p *partition) set(key string, val interface{}) {
	p.Lock()
	defer p.Unlock()

	p.data[key] = val
	return
}
func (p *partition) del(key string) {
	p.Lock()
	defer p.Unlock()

	delete(p.data, key)
}
