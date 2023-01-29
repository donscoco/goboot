package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/donscoco/goboot/container/lrucache"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var LRUServer *Server
var addrP string

func init() {
	flag.StringVar(&addrP, "addr", ":9090", "input addr ")
}

/*
启动三个数据存储节点，
./server -addr=localhost:9090
./server -addr=localhost:9091
./server -addr=localhost:9092

然后在client set和get。观察server的数据情况
*/
func main() {
	flag.Parse()

	LRUServer = CreateServer(addrP, 10)

	LRUServer.Start()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	LRUServer.Stop()
}

type Server struct {
	Data *lrucache.LRUCache

	Api      *ServerApi
	Listener net.Listener
	addr     string

	ctx    context.Context
	cancel context.CancelFunc

	sync.WaitGroup
	sync.Mutex
}

func CreateServer(addr string, cap int) (s *Server) {
	s = &Server{
		Data: lrucache.New(cap),
		addr: addr,
	}

	s.ctx, s.cancel = context.WithCancel(context.TODO())
	return
}

func (s *Server) Start() {

	go func() {
		log.Println("rpc server 启动")
		s.Add(1)
		s.RpcServer()
		s.Done()
		log.Println("rpc server 退出")
	}()

	// 测试查看数据
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			<-ticker.C
			s.Data.ForTest()
		}
	}()

}
func (s *Server) Stop() {
	s.cancel()
	s.Listener.Close()
	s.Wait()
	log.Println("安全退出")
}
func (s *Server) RpcServer() (err error) {

	serverApi := new(ServerApi)
	s.Api = serverApi
	serverApi.server = s
	err = rpc.Register(serverApi)
	if err != nil {
		log.Fatalln(err)
	}

	_, port, _ := net.SplitHostPort(s.addr)
	s.Listener, err = net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("listen on ", s.addr)

	var tempDelay time.Duration

	for {
		conn, err := s.Listener.Accept()
		// 直接抄 net/http 的处理方式
		if err != nil {
			select {
			case <-s.ctx.Done():
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
		}

		go s.serve(conn)
	}

}
func (s *Server) serve(conn net.Conn) {
	s.Add(1)
	rpc.ServeConn(conn)
	s.Done()
}

type ServerApi struct {
	server *Server
}
type ApiReq struct {
	Key string
	Val interface{}
}
type ApiReply struct {
	IsSuccess bool
	Val       interface{}
}

func (s *ServerApi) Get(req ApiReq, reply *ApiReply) error {
	log.Println("调用Get")
	reply.Val, reply.IsSuccess = s.server.Data.Get(req.Key)
	return nil
}
func (s *ServerApi) Set(req ApiReq, reply *ApiReply) error {
	log.Println("调用Set")
	s.server.Data.Set(req.Key, req.Val)
	reply.Val, reply.IsSuccess = s.server.Data.Get(req.Key)
	return nil
}
