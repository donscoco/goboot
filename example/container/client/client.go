package main

import (
	"fmt"
	consistent_hashing "goboot/container/consistent-hashing"
	"log"
	"net"
	"net/rpc"
	"strconv"
	"time"
)

func main() {
	nodes := []string{"localhost:9090", "localhost:9091", "localhost:9092"}
	clientAgent := CreateClient(nodes)

	for i := 0; i < 30; i++ {
		req := ApiReq{
			Key: strconv.Itoa(i),
			Val: i,
		}
		reply := ApiReply{}
		err := clientAgent.Set(req, &reply)
		if err != nil {
			log.Println(err)
		}
	}

	for i := 0; i < 30; i++ {
		req := ApiReq{
			Key: strconv.Itoa(i),
			Val: i,
		}
		reply := ApiReply{}
		err := clientAgent.Get(req, &reply)
		if err != nil {
			log.Println(err)
		}
		if reply.IsSuccess {
			log.Println(reply.Val)
		}
	}

}

const (
	RpcApiGet = "ServerApi.Get"
	RpcApiSet = "ServerApi.Set"
)

type ApiReq struct {
	Key string
	Val interface{}
}
type ApiReply struct {
	IsSuccess bool
	Val       interface{}
}

// demo简单做，不考虑连接池等
type Client struct {
	Locator *consistent_hashing.ConsistentHash

	//Agent []*rpc.Client
	Nodes []string
}

func CreateClient(nodes []string) (c *Client) {
	c = new(Client)
	c.Locator = consistent_hashing.NewConsistentHash()
	c.Nodes = nodes
	for _, addr := range c.Nodes {
		//conn, err := net.Dial("tcp", addr)
		//if err != nil {
		//}
		//cli := rpc.NewClient(conn)
		c.Locator.Add(addr, addr, 8)
	}

	return
}
func (c *Client) Get(req ApiReq, reply *ApiReply) (err error) {
	addr := c.Locator.Get(req.Key).(string)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
	}
	cli := rpc.NewClient(conn)
	defer cli.Close()

	select {
	case <-time.After(1 * time.Second):
		err = fmt.Errorf("call timeout")
	case call := <-cli.Go(RpcApiGet, req, reply, nil).Done:
		err = call.Error
	}
	if err != nil {
		//todo
		return
	}
	return nil
}

// 简单做demo。每次调用就去建立连接
func (c *Client) Set(req ApiReq, reply *ApiReply) (err error) {
	addr := c.Locator.Get(req.Key).(string)

	log.Println("locate ", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	cli := rpc.NewClient(conn)
	defer cli.Close()

	select {
	case <-time.After(1 * time.Second):
		err = fmt.Errorf("call timeout")
	case call := <-cli.Go(RpcApiSet, req, reply, nil).Done:
		err = call.Error
	}
	if err != nil {
		//todo
		return
	}
	return nil
}
