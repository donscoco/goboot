package lfucache

import (
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	lfu := New(5)

	for i := 0; i < 5; i++ {
		lfu.Set(strconv.Itoa(i), i)
	}

	// 0 1 2 频次+1，不会优先被踢出去
	for i := 0; i < 3; i++ {
		lfu.Get(strconv.Itoa(i))
	}

	// 加入 5 6 7，因为 5 频次只有1 加入7的时候容量不够，5被踢出去了
	for i := 5; i < 8; i++ {
		lfu.Set(strconv.Itoa(i), i)
	}

	//lfu.ForTest()
}
