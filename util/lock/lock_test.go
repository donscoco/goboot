package lock

import (
	"github.com/donscoco/goboot/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// 测试redis分布式锁
func TestCreateRedisLocker(t *testing.T) {
	config.ConfigFilePath = "../../config-demo.json"
	conf := config.NewConfiguration(config.ConfigFilePath)

	//locker, err := CreateRedisLocker(conf, "/core/lock/redis", "lockname", 100)
	//if err != nil {
	//	log.Fatal(err)
	//}

	Init(conf, "/core/lock/redis")
	locker := DefaultRedisLocker

	for i := 0; i < 10; i++ {
		go func(n int) {
			for {
				islock, _ := locker.TryLock()
				if islock {
					log.Printf("协程%d 抢占锁\n", n)
					//time.Sleep(5 * time.Second)
					//time.Sleep(20 * time.Second)
					for j := 0; j < 10; j++ {
						log.Printf("协程%d 执行业务逻辑\n", n)
						time.Sleep(1 * time.Second)
					}
					locker.UnLock()
					log.Printf("协程%d 释放锁\n", n)
				}
				time.Sleep(5 * time.Second)
			}
		}(i)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}

// 测试zk分布式锁
func TestCreateZKLocker(t *testing.T) {
	config.ConfigFilePath = "../../config-demo.json"
	conf := config.NewConfiguration(config.ConfigFilePath)

	InitZK(conf, "/core/lock/zk")
	locker := DefaultZKLocker

	// 模拟10个协程/进程/服务器 请求分布式锁,在其他机器也启动观察。
	for i := 0; i < 10; i++ {
		go func(n int) {
			for {
				locker.Lock()
				log.Printf("进程%d 执行业务逻辑", n)
				time.Sleep(3 * time.Second)
				locker.Unlock()
			}
		}(i)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

}
