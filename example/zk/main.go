package main

import (
	"fmt"
	"goboot/config"
	"goboot/coordinator/mzk"
	"goboot/log/mlog"
	"time"
)

func main() {
	config.ConfigFilePath = "./zk.json"
	conf := config.NewConfiguration(config.ConfigFilePath)
	mlog.InitLoggerByConfig(conf, "/core/log")

	c, err := mzk.CreateCoordinator(conf, "/core/coordinator/zookeeper")
	if err != nil {
		fmt.Println(err)
	}
	c.Start()

	n, _ := c.CreateNode("/node1", []byte("data1"), true)
	n, _ = c.GetNode("/node1")

	// watch 一个节点，然后启一个协程去改动这个节点，观察watch情况
	go n.Watch(c.GetContext(), mzk.EventDataChange|mzk.EventChildChange, watchCallback)
	go func() {
		time.Sleep(3 * time.Second)
		n.Set([]byte("data1"))

		time.Sleep(3 * time.Second)
		nc, _ := n.Create("/child1", []byte("data1"), false)

		time.Sleep(3 * time.Second)
		nc.Set([]byte("data3"))

		nc.Remove()
	}()

	for {
		//连接上 后，关闭任意一个节点，只要还有一个节点在，就还能继续处理。
		time.Sleep(1 * time.Second)
		d, err := n.Get()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(string(d))
	}

	c.Stop()
}

func watchCallback(param *mzk.CallbackParam) error {
	if param.EventType == mzk.EventDataChange {
		fmt.Printf("recv DataChange, before change %s\n", string(param.Data.([]byte)))
	}
	if param.EventType == mzk.EventChildChange {
		fmt.Printf("recv ChildChange, before change child : %+v \n", param.Data.([]string))
	}
	return nil
}
