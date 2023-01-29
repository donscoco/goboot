package mkafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/donscoco/goboot/config"
	"github.com/donscoco/goboot/log/mlog"
	"sync"
)

/*
	https://pkg.go.dev/github.com/Shopify/sarama
	https://github.com/Shopify/sarama
*/

var logger = mlog.NewLogger("mq-kafka")

type KafkaProducer struct {
	config struct {
		Name    string
		Brokers []string // 服务器列表
		SASL    struct {
			Enable   bool // 是否启用加密信息，没有密码的服务器不要设置为true，即使不填user和pass也连不上
			User     string
			Password string
		}

		SetLog       bool
		ReturnSuc    bool   // 打印发送成功的信息
		WaitForAll   bool   // 所有kafka节点都要收到消息才继续
		RequiredAcks string //
		Version      string // 驱动应该用什么版本的api与服务器对话，最好和服务器版本保持一致以避免怪问题
	}
	saramaConfig *sarama.Config
	producer     sarama.AsyncProducer
	//input        chan interface{}
	input chan *sarama.ProducerMessage

	sync.WaitGroup
}

func CreateProducer(config *config.Config, path string) (p *KafkaProducer, err error) {
	p = new(KafkaProducer)

	err = config.GetByScan(path, &p.config)
	if err != nil {
		return nil, err
	}

	sconf := sarama.NewConfig()
	switch p.config.RequiredAcks {
	case "NoResponse":
		sconf.Producer.RequiredAcks = sarama.NoResponse
	case "WaitForLocal":
		sconf.Producer.RequiredAcks = sarama.WaitForLocal
	case "WaitForAll":
		sconf.Producer.RequiredAcks = sarama.WaitForAll
	default:
		sconf.Producer.RequiredAcks = sarama.WaitForAll
	}

	if p.config.ReturnSuc {
		sconf.Producer.Return.Successes = true
	}
	if p.config.SASL.Enable {
		sconf.Net.SASL.Enable = true
		sconf.Net.SASL.User = p.config.SASL.User
		sconf.Net.SASL.Password = p.config.SASL.Password
	}
	if p.config.Version != "" {
		kafkaVersion, err := sarama.ParseKafkaVersion(p.config.Version)
		if err == nil {
			sconf.Version = kafkaVersion
			logger.Debugf("Kafka 版本配置为: %s", kafkaVersion)
		} else {
			logger.Errorf("错误的 Kafka 版本配置: %s", p.config.Version)
		}
	}

	if p.config.SetLog {
		sarama.Logger = newSaramaLogger()
	}

	p.saramaConfig = sconf
	p.producer, err = sarama.NewAsyncProducer(p.config.Brokers, p.saramaConfig)
	if err != nil {
		return nil, err
	}
	p.input = make(chan *sarama.ProducerMessage)
	return
}
func (p *KafkaProducer) Start() {

	go func() {
		p.Add(1)
		p.Worker()
		p.Done()
	}()

}
func (p *KafkaProducer) Worker() {

	for {
		select {
		case msg, ok := <-p.input:
			if !ok {
				// 收到关闭通知，关闭下游
				p.producer.Close()
				return
			}
			p.producer.Input() <- msg

		case msg, ok := <-p.producer.Successes():
			if !ok {
				// todo
				return
			}
			logger.Debugf("%s成功发送消息: %s", p, msg)

		case err, ok := <-p.producer.Errors():
			if !ok {
				// todo
				return
			}
			logger.Errorf("%s出错: %s", p, err)

		}

	}
}
func (p *KafkaProducer) Stop() {
	close(p.input)

	p.Wait()
}
func (p *KafkaProducer) Input() chan *sarama.ProducerMessage {
	return p.input
}
func (p *KafkaProducer) String() string {
	return fmt.Sprintf("生产者(name=%s, brokers=%v)", p.config.Name, p.config.Brokers)
}

func NewMessage(topic string, data []byte) (msg *sarama.ProducerMessage) {
	msg = &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	return
}
