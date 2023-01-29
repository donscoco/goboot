package mredis

import (
	"fmt"
	"github.com/donscoco/goboot/config"
	"github.com/donscoco/goboot/log/mlog"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestCreateRedisMQ(t *testing.T) {

	config.ConfigFilePath = "../../config-demo.json"
	conf := config.NewConfiguration(config.ConfigFilePath)

	mlog.InitLoggerByConfig(conf, "/core/log")

	rps, err := CreateRedisMQ(conf, "/core/mq/redis")
	if err != nil {
		log.Fatal(err)
	}
	rps.Subscribe("ironhead-mq-redis", func(topic string, msg []byte) error {
		fmt.Println("recv:", string(msg))
		return nil
	})
	(rps).Start()
	go func() {
		for i := 0; i < 10; i++ {
			(rps).Publish("ironhead-mq-redis", "sendedMsg-"+strconv.Itoa(i))
			time.Sleep(1)
		}
	}()

	time.Sleep(3 * time.Second)
	(rps).Stop(nil)

}
