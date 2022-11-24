package mzk

import (
	"context"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
)

type Node struct {
	path    string
	absPath string
	stat    *zk.Stat

	zk *ZookeeperCoordinator
}

func (n *Node) Create(path string, data []byte, isPersisted bool) (*Node, error) {
	return n.zk.CreateNode(n.path+path, data, isPersisted)
}

func (n *Node) CreateIfNotExist(path string, data []byte, isPersisted bool) (*Node, error) {
	return n.zk.CreateNodeIfNotExist(n.path+path, data, isPersisted)
}

func (n *Node) Get() ([]byte, error) {
	data, stat, err := n.zk.conn.Get(n.absPath)
	if err == nil {
		n.stat = stat
	}
	return data, err
}
func (n *Node) Set(data []byte) (err error) {
	stat, err := n.zk.conn.Set(n.absPath, data, n.stat.Version)
	if err == nil {
		n.stat = stat
	}
	return
}
func (n *Node) GetChildren() (cNodes []string, err error) {
	cNodes, _, err = n.zk.conn.Children(n.absPath)
	return
}
func (n *Node) GetChildrenNodes() (cNodes []*Node, err error) {
	cNodesName, _, err := n.zk.conn.Children(n.absPath)
	for _, path := range cNodesName {
		cnode, err := n.zk.GetNode(n.path + "/" + path)
		if err != nil {
			return nil, err
		}
		cNodes = append(cNodes, cnode)
	}
	return
}
func (n *Node) Remove() error {
	return n.zk.conn.Delete(n.absPath, n.stat.Version)
}
func (n *Node) Refresh() error {
	exist, stat, err := n.zk.conn.Exists(n.absPath)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("path[%s] not exist", n.absPath)
	}
	n.stat = stat
	return nil
}
func (n *Node) Watch(context context.Context, watchEvent int, callback func(*CallbackParam) error) (err error) {
	conn := n.zk.conn

	var stat *zk.Stat

	var ExistEvent <-chan zk.Event
	var DataEvent <-chan zk.Event
	var ChildEvent <-chan zk.Event
	var exist bool
	var data []byte
	var child []string

	if watchEvent&EventCreate > 0 || watchEvent&EventDelete > 0 {
		exist, stat, ExistEvent, err = conn.ExistsW(n.absPath)
		if err != nil {
			return err
		}
	}

	if watchEvent&EventDataChange > 0 {
		data, stat, DataEvent, err = conn.GetW(n.absPath)
		if err != nil {
			return err
		}
	}
	if watchEvent&EventChildChange > 0 {
		child, stat, ChildEvent, err = conn.ChildrenW(n.absPath)
		if err != nil {
			return err
		}
	}
	n.stat = stat

	for {
		select {
		case e := <-ExistEvent:
			// 检查
			if e.Type != zk.EventNodeCreated || e.Type != zk.EventNodeDeleted {
				// todo
			}

			// 参数
			et := new(CallbackParam)
			if e.Type == zk.EventNodeCreated {
				et.EventType = EventCreate
			}
			if e.Type == zk.EventNodeDeleted {
				et.EventType = EventDelete
			}
			et.Data = exist // data是数据改变前的

			// 处理
			err = callback(et)
			if err != nil {
				return err
			}

			// watch
			exist, stat, ExistEvent, err = conn.ExistsW(n.absPath)
			if err != nil {
				return err
			}
		case e := <-DataEvent:
			if e.Type != zk.EventNodeDataChanged {
				// todo
			}

			et := new(CallbackParam)
			et.EventType = EventDataChange
			et.Data = data // data是数据改变前的

			err = callback(et)
			if err != nil {
				return nil
			}

			data, stat, DataEvent, err = conn.GetW(n.absPath)
			if err != nil {
				return nil
			}
			n.stat = stat
		case e := <-ChildEvent:
			// 检查
			if e.Type != zk.EventNodeChildrenChanged {
				// todo
			}

			// 参数
			et := new(CallbackParam)
			et.EventType = EventChildChange
			et.Data = child // data是数据改变前的

			// 处理
			err = callback(et)
			if err != nil {
				return err
			}

			// watch
			child, stat, ChildEvent, err = conn.ChildrenW(n.absPath)
			if err != nil {
				return err
			}
			n.stat = stat

		case <-context.Done():
			// 关闭下游
			return
		}
	}

	return nil
}

type CallbackParam struct {
	EventType int
	Data      interface{}
}
