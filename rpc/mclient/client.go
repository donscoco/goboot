package mclient

import (
	"context"
	"encoding/json"
	"goboot/config"
	"goboot/coordinator/mzk"
	"goboot/rpc/mserver"
	"sync"
)

type Config struct {
	Node          string
	CallTimeoutMs int // 请求超时时间,单位:毫秒
	DialTimeoutMs int // 建立连接超时时间,单位:毫秒
	MaxConns      int // 最多连接数
	MinConns      int // 最少连接数
	Retries       int // 重试次数
}
type RpcClient struct {
	Config Config

	// server 对应的coordinator 上的信息
	serverNode  *mzk.Node
	coordinator *mzk.ZookeeperCoordinator

	clientAgent map[string]*ClientAgent
	balance     []*ClientAgent
	cursor      int // 均衡轮询的游标

	context context.Context
	cancel  context.CancelFunc

	sync.WaitGroup
	sync.Mutex
}

func CreateRpcClient(config *config.Config, path string, coordinator *mzk.ZookeeperCoordinator) (c *RpcClient, err error) {
	c = new(RpcClient)
	err = config.GetByScan(path, &c.Config)
	if err != nil {
		return nil, err
	}
	if c.Config.Retries == 0 {
		c.Config.Retries = 1
	}

	if coordinator != nil {
		c.coordinator = coordinator
	}
	c.clientAgent = make(map[string]*ClientAgent)
	c.balance = make([]*ClientAgent, 0, 16)

	c.context, c.cancel = context.WithCancel(context.TODO())

	return c, nil
}

func (c *RpcClient) Start() (err error) {

	// watch 节点
	root, err := c.coordinator.GetNode("")
	if err != nil {
		return err
	}
	serverNode, err := root.CreateIfNotExist(c.Config.Node, nil, true)
	if err != nil {
	}

	c.serverNode = serverNode

	err = c.updateClientAgent()
	if err != nil {
		return
	}

	go func() {
		c.Add(1)
		c.serverNode.Watch(c.context, mzk.EventChildChange|mzk.EventDataChange|mzk.EventDelete, c.handler)
		c.Done()
	}()

	return
}
func (c *RpcClient) Stop() (err error) {
	c.cancel()

	for _, ca := range c.clientAgent {
		ca.Close()
	}

	c.Wait()
	return
}

func (c *RpcClient) Call(method string, args interface{}, reply interface{}) (err error) {

	for i := 0; i < c.Config.Retries; i++ {
		// 获得agent的一个client , 调用
		ca, err := c.GetClientAgent()
		if err != nil {
			continue
		}
		err = ca.Call(method, args, reply, int64(c.Config.CallTimeoutMs))
		if err != nil {
			continue
		} else {
			break
		}
	}
	return err

}
func (c *RpcClient) GetClientAgent() (ca *ClientAgent, err error) {

	len := len(c.balance)
	if len == 0 {
		return nil, ErrNoServers
	}

	c.Lock()
	defer c.Unlock()

	c.cursor = (c.cursor + 1) % len
	return c.balance[c.cursor], nil
}

func (c *RpcClient) updateClientAgent() (err error) {
	//获取最新的服务uri
	newAddrs := make([]string, 0, 16)
	cNodes, err := c.serverNode.GetChildrenNodes()
	if err != nil {
		return err
	}

	for _, cn := range cNodes {
		data, err := cn.Get()
		if err != nil {
			return err
		}
		serverInfo := &mserver.State{}
		err = json.Unmarshal(data, serverInfo)
		if err != nil {
			return err
		}
		newAddrs = append(newAddrs, serverInfo.Addr)
	}

	//更新agent
	newAgent := make(map[string]*ClientAgent)
	newBalance := make([]*ClientAgent, 0, 16)
	for _, addr := range newAddrs {
		agentp, exist := c.clientAgent[addr]
		if exist {
			newAgent[addr] = agentp
			newBalance = append(newBalance, agentp)
		} else { // 没有就创建新的agent
			agentp, err = CreateClientAgent(addr, c.Config.DialTimeoutMs, c.Config.MaxConns, c.Config.MinConns)
			if err != nil {
				return err
			}
			newAgent[addr] = agentp
			newBalance = append(newBalance, agentp)
		}
	}
	oldAgent := c.clientAgent
	c.clientAgent = newAgent
	c.balance = newBalance

	//关闭旧agent
	for addr, agent := range oldAgent {
		_, exist := newAgent[addr]
		if exist {
			continue
		}
		agent.Close()
	}
	return
}

func (c *RpcClient) handler(p *mzk.CallbackParam) (err error) {

	// todo 适配

	if p.EventType == mzk.EventChildChange {
		// fixme ; 前面server 直接使用 addr 作为节点名。后续修改这里需要加个转化解析node的值，拿到addr
		//cNodes, err := c.serverNode.GetChildren()

		//获取最新的服务uri
		//newAddrs := make([]string, 0, 16)
		//cNodes, err := c.serverNode.GetChildrenNodes()
		//if err != nil {
		//}
		//for _, cn := range cNodes {
		//	data, err := cn.Get()
		//	if err != nil {
		//		return err
		//	}
		//	serverInfo := &mserver.State{}
		//	err = json.Unmarshal(data, serverInfo)
		//	if err != nil {
		//		return err
		//	}
		//	newAddrs = append(newAddrs, serverInfo.Addr)
		//}
		//
		////更新agent
		//newAgent := make(map[string]*ClientAgent)
		//newBalance := make([]*ClientAgent, 0, 16)
		//for _, addr := range newAddrs {
		//	agentp, exist := c.clientAgent[addr]
		//	if exist {
		//		newAgent[addr] = agentp
		//		newBalance = append(newBalance, agentp)
		//	} else { // 没有就创建新的agent
		//		agentp, err = CreateClientAgent(addr, int64(c.Config.DialTimeoutSec))
		//		if err != nil {
		//			return err
		//		}
		//		newAgent[addr] = agentp
		//		newBalance = append(newBalance, agentp)
		//	}
		//}
		//oldAgent := c.clientAgent
		//c.clientAgent = newAgent
		//c.balance = newBalance
		//
		////关闭旧agent
		//for addr, agent := range oldAgent {
		//	_, exist := newAgent[addr]
		//	if exist {
		//		continue
		//	}
		//	agent.Close()
		//}
		err = c.updateClientAgent()
		if err != nil {
			return
		}
	}
	if p.EventType == mzk.EventDataChange {
	}
	if p.EventType == mzk.EventCreate {
	}
	if p.EventType == mzk.EventDelete {
		// todo 关闭所有的client，因为serverNode已经被关闭说明没有服务提供了
	}
	return nil
}
