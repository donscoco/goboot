package main

import (
	"fmt"
	"github.com/donscoco/goboot/config"
	"github.com/donscoco/goboot/coordinator/mzk"
	"github.com/donscoco/goboot/log/mlog"
	"github.com/donscoco/goboot/rpc/mclient"
	"time"
)

func main() {
	//client, err := rpc.Dial("tcp", "localhost:9090")
	//if err != nil {
	//	log.Fatal("dialing:", err)
	//}
	//
	//var reply string
	//err = client.Call("HelloService.Hello", "hello", &reply)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println(reply)

	config.ConfigFilePath = "./rpc-client.json"
	conf := config.NewConfiguration(config.ConfigFilePath)
	mlog.InitLoggerByConfig(conf, "/core/log")

	c, err := mzk.CreateCoordinator(conf, "/core/coordinator/zookeeper")
	if err != nil {
		fmt.Println(err)
	}
	c.Start()

	client, _ := mclient.CreateRpcClient(conf, "/core/rpc/client", c)
	err = client.Start()
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(10 * time.Second)

	for i := 0; i < 10; i++ {
		go func() {
			var reply string
			err = client.Call("HelloService.Hello", "hello", &reply)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(reply)
		}()
	}
	var reply string
	err = client.Call("HelloService.Hello", "hello", &reply)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(reply)
	c.Stop()

}
