package mserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/donscoco/goboot/config"
	"github.com/donscoco/goboot/coordinator/mzk"
	"github.com/donscoco/goboot/util"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Config struct {
	Addr         string
	HeartbeatSec int // 心跳间隔
	Prefix       string

	Node string // 节点名称
}
type State struct {
	Name string
	Addr string

	Data string

	StartTime int64 // 服务开始时间
	Heartbeat int64 // 最后的心跳时间
}
type RpcServer struct {
	// 配置
	Config Config
	// 状态信息，用于保存在coodinator上
	State State

	listener net.Listener

	// server 对应的coordinator 上的信息
	serverNode *mzk.Node
	node       *mzk.Node //本服务对应的节点

	// 一致性服务的连接 // todo 将 coordinator 改成接口
	coordinator *mzk.ZookeeperCoordinator

	context context.Context
	cancel  context.CancelFunc
	sync.WaitGroup
}

// 创建
func CreateRpcServer(config *config.Config, path string, coordinator *mzk.ZookeeperCoordinator, handler interface{}) (s *RpcServer, err error) {

	s = new(RpcServer)
	err = config.GetByScan(path, &s.Config)
	if err != nil {
		return nil, err
	}

	err = rpc.Register(handler)
	if err != nil {
		return nil, err
	}

	if coordinator != nil {
		s.coordinator = coordinator
	}

	// 如果配置文件没有指定ip则自己查找默认的子网ip 192.168 //fixme 兼容处理 k8s下的网络
	host, _, _ := net.SplitHostPort(s.Config.Addr)
	if len(host) == 0 {
		host = util.GetIpv4_192_168()
	}

	s.State = State{
		Name:      s.Config.Node,
		Addr:      host + s.Config.Addr,
		Data:      "",
		StartTime: time.Now().Unix(),
		Heartbeat: 0,
	}

	s.context, s.cancel = context.WithCancel(context.TODO())

	return
}

// start
func (s *RpcServer) Start() (err error) {

	_, port, _ := net.SplitHostPort(s.Config.Addr) // 没必要指定ip来监听
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		// todo
		return err
	}

	go func() {
		s.Add(1)
		s.AcceptRoutine()
		s.Done()
	}()

	// 服务已经启动，注册到协调服务
	err = s.Register()
	if err != nil {
		return err
	}
	// 启动心跳
	go func() {
		s.Add(1)
		s.Heartbeat()
		s.Done()
	}()

	return
}

func (s *RpcServer) AcceptRoutine() (err error) {

	var tempDelay time.Duration
	for {

		conn, err := s.listener.Accept()
		// 直接仿照 net/http 包的异常处理
		if err != nil {
			select {
			case <-s.context.Done():
				return err
			default:
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		go s.serve(conn)

	}

}

// stop
func (s *RpcServer) Stop() {
	s.cancel()

	s.listener.Close()
	//s.coordinator.Stop()

	s.Wait()
	fmt.Println("exit success")
}

// 注册
func (s *RpcServer) Register() (err error) {
	if s.coordinator == nil {
		return
	}

	// 判断节点
	//exist, err := s.coordinator.Exist(s.Config.Node)
	//if err != nil {
	//	// todo
	//	return err
	//}
	//if exist {
	//	return fmt.Errorf("node exist")
	//}

	// 获取 根 节点
	root, err := s.coordinator.GetNode("")
	if err != nil {
		return err
	}

	// 创建/获取服务节点， 第一个启动的服务会去创建对应的服务节点
	serverNode, err := root.CreateIfNotExist(s.Config.Node, nil, true)
	if err != nil {
		return err
	}

	// 创建server对应节点
	data, err := json.Marshal(s.State)
	if err != nil {
		return err
	}
	node, err := serverNode.Create("/"+s.Config.Prefix+s.State.Addr, data, false) // todo 现在是直接使用addr 作为节点名，后续考虑修改
	if err != nil {
		return err
	}

	s.serverNode = serverNode
	s.node = node

	return nil
}

// 心跳
func (s *RpcServer) Heartbeat() (err error) {
	if s.coordinator == nil {
		return
	}
	ticker := time.NewTicker(time.Duration(s.Config.HeartbeatSec) * time.Second)
	for {
		select {
		case <-s.context.Done():
			// 收到退出通知，关闭下游
			return
		case <-ticker.C:

			s.State.Heartbeat = time.Now().Unix()
			data, _ := json.Marshal(s.State) // 前面marshal没有错，这里marshal肯定也不会错直接不用管err

			err = s.node.Set(data)
			if err != nil {
				// todo 重试次数超时报警等机制
				// 连接不上coordinator，尝试重新注册
				err = s.Register()
				if err != nil {
					fmt.Println(err)
					//return
				}
			}
		}
	}
	return nil
}

func (s *RpcServer) serve(conn net.Conn) {
	s.Add(1)            // 防止处理一半 stop 退出被关掉，让 stop() 等一下;上游已经关闭listener，不用担心一直add导致这里退不出去
	rpc.ServeConn(conn) // ServeConn 会去close 掉 conn
	s.Done()
}
