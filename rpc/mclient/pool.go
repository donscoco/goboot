package mclient

import (
	"errors"
	"io"
	"net"
	"net/rpc"
	"sync"
	"time"
)

var (
	ErrNoServers    = errors.New("没有可用服务器节点")
	ErrNoConnection = errors.New("没有可用连接")
	ErrShutdown     = errors.New("连接关闭")
	ErrTimeout      = errors.New("请求超时")
	ErrEOF          = errors.New("EOF")
)

const (
	defaultMaxConn = 100
	defaultMinConn = 1
)

// 用于连接每个服务节点
type ClientAgent struct {
	client *rpc.Client // 如果只有一个client 并发性能不高,多个协程会在同一个client里面排队
	//todo 连接池

	serverAddr  string
	dialTimeout int
	maxConn     int
	minConn     int
	pool        []*rpc.Client

	sync.Mutex
}

func CreateClientAgent(serverAddr string, dialTimeout, maxConn, minConn int) (ca *ClientAgent, err error) {
	ca = new(ClientAgent)
	ca.serverAddr = serverAddr
	if maxConn != 0 {
		ca.maxConn = maxConn
	} else {
		ca.maxConn = defaultMaxConn
	}
	if minConn != 0 {
		ca.minConn = minConn
	} else {
		ca.minConn = defaultMinConn
	}
	ca.dialTimeout = dialTimeout

	conn, err := net.DialTimeout("tcp", ca.serverAddr, time.Duration(ca.dialTimeout)*time.Millisecond)
	if err != nil {
		return nil, err
	}
	ca.client = rpc.NewClient(conn)

	return
}

func (ca *ClientAgent) Call(method string, args interface{}, reply interface{}, timeout int64) (err error) {

	var client *rpc.Client
	client, err = ca.acquireForce()
	if err != nil {
		return
	}

	// 超时处理
	if timeout > 0 {
		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			err = ErrTimeout
		case call := <-client.Go(method, args, reply, nil).Done:
			err = call.Error
		}
	} else {
		err = client.Call(method, args, reply)
	}

	switch err {
	case ErrTimeout:
		// 超时的连接不要了
		client.Close()
	case rpc.ErrShutdown:
		err = ErrShutdown
		client.Close()
	case io.EOF:
		err = ErrEOF
		client.Close()
	case nil:
		break
	default:
		if ne, ok := err.(net.Error); ok && ne.Timeout() { // 如果是超时错误，弃掉这个连接
			err = ErrTimeout
			client.Close()
		}
	}
	ca.release(client)
	return
}

// 无论如何，都要返回新的连接，可能超出maxconn
func (ca *ClientAgent) acquireForce() (c *rpc.Client, err error) {
	c = ca.acquire()
	if c == nil {
		conn, err := net.DialTimeout("tcp", ca.serverAddr, time.Duration(ca.dialTimeout)*time.Millisecond)
		if err != nil {
			return nil, err
		}
		c = rpc.NewClient(conn)
	}
	return
}

//////////////////////  对 pool 操作涉及 lock 的处理函数 //////////////////////

func (ca *ClientAgent) Close() {
	ca.Lock()
	defer ca.Unlock()

	for _, c := range ca.pool {
		c.Close()
	}
	ca.pool = nil

}
func (ca *ClientAgent) acquire() (c *rpc.Client) {
	ca.Lock()
	defer ca.Unlock()

	if len(ca.pool) > 0 {
		c = ca.pool[0]
		ca.pool = ca.pool[1:]
	}

	return

}
func (ca *ClientAgent) release(c *rpc.Client) {
	ca.Lock()
	defer ca.Unlock()

	if len(ca.pool) > ca.minConn {
		c.Close()
	} else {
		ca.pool = append(ca.pool, c)
	}

	return
}
