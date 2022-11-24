package main

import (
	"fmt"
	"goboot/config"
	"goboot/coordinator/mzk"
	"goboot/log/mlog"
	"goboot/rpc/mclient"
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

}
