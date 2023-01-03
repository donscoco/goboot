package redis

import (
	"goboot/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestCreateRedisLocker(t *testing.T) {
	config.ConfigFilePath = "../../../config-demo.json"
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
