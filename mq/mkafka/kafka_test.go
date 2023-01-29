package mkafka

import (
	"github.com/donscoco/goboot/config"
	"github.com/donscoco/goboot/log/mlog"
	"log"
	"testing"
	"time"
)

func TestKafkaProducer_Start(t *testing.T) {
	config.ConfigFilePath = "../../config-demo.json"
	conf := config.NewConfiguration(config.ConfigFilePath)

	mlog.InitLoggerByConfig(conf, "/core/log")

	producer, err := CreateProducer(conf, "/core/mq/kafka/producer-config")
	if err != nil {
		log.Fatalf("create producer err %s", err.Error())
	}
	producer.Start()
	producer.Input() <- NewMessage("domark-test", []byte("test1"))
	producer.Stop()

	consumer, err := CreateConsumer(conf, "/core/mq/kafka/consumer-config")
	if err != nil {
		log.Fatalf("create consumer err %s", err.Error())
	}
	consumer.Start()
	go func() {
		for {
			select {
			case msg := <-consumer.Output():
				log.Println("recv:", msg)
			}
		}

	}()
	time.Sleep(1 * time.Second)
	consumer.Stop()

}
