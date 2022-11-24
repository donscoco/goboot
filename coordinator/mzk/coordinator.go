package mzk

import (
	"context"
	"errors"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"goboot/config"
	"goboot/log/mlog"
	"sync"
	"time"
)

/*
	https://pkg.go.dev/github.com/samuel/go-zookeeper/zk
*/

var logger = mlog.NewLogger("coordinator-zk")

type ZookeeperCoordinator struct {
	config struct {
		Addrs   []string
		Timeout int
		Path    string
	}
	conn   *zk.Conn
	events <-chan zk.Event

	//closeAction chan struct{}
	//closeChild map[string]*Node
	context context.Context
	cancel  context.CancelFunc

	sync.WaitGroup
}

func CreateCoordinator(config *config.Config, path string) (zkc *ZookeeperCoordinator, err error) {
	zkc = new(ZookeeperCoordinator)
	err = config.GetByScan(path, &zkc.config)
	if err != nil {
		return nil, err
	}

	zkc.conn, zkc.events, err = zk.Connect(zkc.config.Addrs, time.Duration(zkc.config.Timeout)*time.Second)
	if err != nil {
		return nil, err
	}

	// todo

	// 初始化
	//zkc.closeAction = make(chan struct{}, 1)
	zkc.context, zkc.cancel = context.WithCancel(context.TODO())

	return zkc, nil
}

func (zkc *ZookeeperCoordinator) Start() (err error) {

	//检查
	isexist, _, err := zkc.conn.Exists(zkc.config.Path)
	if err != nil {
		return err
	}
	if isexist {
		return fmt.Errorf("path[%s]  exist", zkc.config.Path)
	}
	var flags int32 = 0
	//flags = zk.FlagEphemeral
	//创建对应的 node
	path, err := zkc.conn.Create(zkc.config.Path, nil, flags, zk.WorldACL(zk.PermAll))
	if err != nil {
		return err
	}
	if path != zkc.config.Path {
		return errors.New("path error ")
	}

	go func() {
		zkc.Add(1)
		zkc.Worker()
		zkc.Done()
	}()

	return nil

}
func (zkc *ZookeeperCoordinator) Stop() {
	zkc.cancel()

	zkc.Wait()
}

func (zkc *ZookeeperCoordinator) Worker() {

	for {
		select {
		case event, ok := <-zkc.events:
			if !ok {
				//todo
			}
			logger.Debugf("recv :%+v", event)
			// todo 收到 对应node的 事件，做相应处理
		case <-zkc.context.Done():
			// 收到关闭通知，关闭下游
			zkc.conn.Close()
			return
		}
	}

}

func (zkc *ZookeeperCoordinator) CreateNode(path string, data []byte, isPersisted bool) (n *Node, err error) {

	n = new(Node)
	n.path = path
	n.absPath = zkc.config.Path + n.path

	// 默认是临时节点
	var flags int32
	flags = zk.FlagEphemeral
	// 持久化节点
	if isPersisted {
		flags = 0
	}
	_, err = zkc.conn.Create(n.absPath, data, flags, zk.WorldACL(zk.PermAll))
	if err != nil {
		return nil, err
	}
	isExist, stat, err := zkc.conn.Exists(n.absPath)
	if err != nil {
		return nil, err
	}
	if !isExist {
		return nil, fmt.Errorf("path[%s] not exist", path)
	}

	n.stat = stat
	n.zk = zkc
	return
}

func (zkc *ZookeeperCoordinator) GetNode(path string) (n *Node, err error) {
	n = new(Node)
	n.path = path
	n.absPath = zkc.config.Path + n.path

	isExist, stat, err := zkc.conn.Exists(n.absPath)
	if err != nil {
		return nil, err
	}
	if !isExist {
		return nil, fmt.Errorf("path[%s] not exist", path)
	}

	n.stat = stat
	n.zk = zkc
	return

}
func (zkc *ZookeeperCoordinator) CreateNodeIfNotExist(path string, data []byte, isPersisted bool) (n *Node, err error) {
	isExist, stat, err := zkc.conn.Exists(zkc.config.Path + path)
	if err != nil {
		return nil, err
	}
	if isExist {
		n = new(Node)
		n.path = path
		n.absPath = zkc.config.Path + path
		n.stat = stat
		n.zk = zkc
		//n, err = zkc.GetNode(zkc.config.Path + n.path)
	} else {
		n, err = zkc.CreateNode(path, data, isPersisted)
	}
	return
}

func (zkc *ZookeeperCoordinator) Exist(path string) (bool, error) {
	isExist, _, err := zkc.conn.Exists(zkc.config.Path + path)
	if err != nil {
		return false, err
	}
	return isExist, nil
}
func (zkc *ZookeeperCoordinator) Lookup(path string) (n *Node) {
	n = new(Node)
	n.path = path
	n.absPath = zkc.config.Path + n.path
	n.zk = zkc
	return
}

func (zkc *ZookeeperCoordinator) GetContext() context.Context {
	return zkc.context
}
func (zkc *ZookeeperCoordinator) GetConn() *zk.Conn {
	return zkc.conn
}

// todo 递归创建节点
