package consistent_hashing

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type ConsistentHash struct {
	circle hashcode
	nodes  map[uint32]interface{}
	vnodes map[string]int // 每个节点对应的虚拟节点数量

	sync.Mutex
}

func NewConsistentHash() (c *ConsistentHash) {
	c = new(ConsistentHash)
	c.circle = make([]uint32, 0, 16)
	c.nodes = make(map[uint32]interface{}, 16)
	c.vnodes = make(map[string]int, 16)
	return
}

// 添加节点到 hash 环
func (c *ConsistentHash) Add(name string, node interface{}, virtualNum int) {
	c.Lock()
	defer c.Unlock()

	if virtualNum == 0 { // 每个节点在映射出多个虚拟节点，为了在哈希环上分布均匀，至少一个节点
		virtualNum = 1
	}

	for i := 0; i < virtualNum; i++ {
		h := hashKey(name + strconv.Itoa(i))
		c.circle = append(c.circle, h)
		sort.Sort(c.circle)
		c.nodes[h] = node
	}
	c.vnodes[name] = virtualNum
}

func (c *ConsistentHash) Remove(name string) {
	c.Lock()
	defer c.Unlock()

	virtualNum := c.vnodes[name]
	for i := 0; i < virtualNum; i++ {
		h := hashKey(name + strconv.Itoa(i))
		newCircle := make([]uint32, 0, len(c.circle))
		for _, hashcode := range c.circle {
			if hashcode == h {
				continue
			}
			newCircle = append(newCircle, hashcode) // 这里原本就有序，不需要重新排序
		}
		c.circle = newCircle
		delete(c.nodes, h)
	}
	delete(c.vnodes, name)

	//sort.Sort(c.circle)
}

func (c *ConsistentHash) Get(key string) (node interface{}) {
	c.Lock()
	defer c.Unlock()

	hc := hashKey(key)
	i := 0
	for ; i < len(c.circle); i++ {
		if c.circle[i] > hc {
			break
		}
	}
	if i == len(c.circle) {
		i = 0
	}
	return c.nodes[c.circle[i]]

}

// 实现排序接口
type hashcode []uint32

func (h hashcode) Len() int           { return len(h) }
func (h hashcode) Less(i, j int) bool { return h[i] < h[j] }
func (h hashcode) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func hashKey(key string) uint32 {
	if len(key) < 64 {
		var scratch [64]byte
		copy(scratch[:], key)
		return crc32.ChecksumIEEE(scratch[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}
