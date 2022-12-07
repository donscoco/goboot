package consistent_hashing

import (
	"fmt"
	"testing"
)

func TestNewConsistentHash(t *testing.T) {
	type Node struct {
		Name string
		Addr string
	}

	var n1 = Node{Name: "n91", Addr: "192.168.2.1:9091"}
	var n2 = Node{Name: "n92", Addr: "192.168.2.1:9092"}
	var n3 = Node{Name: "n93", Addr: "192.168.2.1:9093"}

	var ch = NewConsistentHash()
	ch.Add(n1.Name, n1, 1)
	ch.Add(n2.Name, n2, 1)
	ch.Add(n3.Name, n3, 1)

	var key1 = "abc"
	var key2 = "def"
	var key3 = "ghi"

	fmt.Printf("node1 hashcode:%d \n", hashKey(n1.Name+"0"))
	fmt.Printf("node2 hashcode:%d \n", hashKey(n2.Name+"0"))
	fmt.Printf("node3 hashcode:%d \n", hashKey(n3.Name+"0"))

	ret1 := ch.Get(key1)
	fmt.Printf("get key1:%d, return node:%d \n", hashKey(key1), hashKey(ret1.(Node).Name+"0"))
	ret2 := ch.Get(key2)
	fmt.Printf("get key2:%d, return node:%d \n", hashKey(key2), hashKey(ret2.(Node).Name+"0"))
	ret3 := ch.Get(key3)
	fmt.Printf("get key3:%d, return node:%d \n", hashKey(key3), hashKey(ret3.(Node).Name+"0"))

}
