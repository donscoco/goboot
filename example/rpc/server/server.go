package main

import (
	"fmt"
	"github.com/donscoco/goboot/config"
	"github.com/donscoco/goboot/coordinator/mzk"
	"github.com/donscoco/goboot/log/mlog"
	"github.com/donscoco/goboot/rpc/mserver"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//rpc.RegisterName("HelloService", new(HelloService))
	//
	//listener, err := net.Listen("tcp", ":9090")
	//if err != nil {
	//	log.Fatal("ListenTCP error:", err)
	//}
	//
	//conn, err := listener.Accept()
	//if err != nil {
	//	log.Fatal("Accept error:", err)
	//}
	//
	//rpc.ServeConn(conn)

	config.ConfigFilePath = "./rpc-server.json"
	conf := config.NewConfiguration(config.ConfigFilePath)
	mlog.InitLoggerByConfig(conf, "/core/log")
	//s, _ := mserver.CreateRpcServer(conf, "/core/rpc/server", nil, new(HelloService)) //先不要注册到 coordinator
	//s.Start()

	c, err := mzk.CreateCoordinator(conf, "/core/coordinator/zookeeper")
	if err != nil {
		fmt.Println(err)
	}
	c.Start()
	s, _ := mserver.CreateRpcServer(conf, "/core/rpc/server", c, new(HelloService)) //先不要注册到 coordinator
	err = s.Start()
	if err != nil {
		fmt.Println(err)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	s.Stop()
}

type HelloService struct{}

func (p *HelloService) Hello(request string, reply *string) error {
	time.Sleep(5 * time.Second) // 测试安全退出
	*reply = "hello:" + request
	return nil
}
