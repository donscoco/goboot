package lrucache

import (
	"strconv"
	"testing"
)

func TestLRUCache_Set(t *testing.T) {

	lru := New(10)
	//lru.Set("key-1", 10)
	//fmt.Println(lru.Get("key-1"))

	for i := 0; i < 10; i++ {
		lru.Set(strconv.Itoa(i), i)
	}
	target := []string{"9", "1", "8", "2", "7", "3", "6", "4", "5", "0"}
	for i := 0; i < 10; i++ {
		lru.Get(target[i])
	}

	// 054646372819
	lru.ForTest()
	//
	//fmt.Println("----------")

	lru = New(10)
	for i := 0; i < 10; i++ {
		lru.Set(strconv.Itoa(i), i)
	}
	target = []string{"0", "5", "1", "9", "2", "8"}
	for i := 0; i < len(target); i++ {
		lru.Get(target[i])
	}
	for i := 10; i < 15; i++ {
		lru.Set(strconv.Itoa(i), i)
	}
	//// [14 13 12 11 10 8 2 9 1 5] 0 8 7 6 4 3
	lru.ForTest()
}
