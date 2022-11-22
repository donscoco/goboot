package mkafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"goboot/config"
	"goboot/log/mlog"
	"os"
	"sync"
	"time"
)

type KafkaConsumer struct {
	config struct {
		Name    string
		Brokers []string // 服务器列表
		SASL    struct {
			Enable   bool // 是否启用加密信息，没有密码的服务器不要设置为true，即使不填user和pass也连不上
			User     string
			Password string
		}
		ClientId   string
		GroupId    string   // 消费组
		Topics     []string // topic
		AutoCommit struct {
			Enable   bool
			Interval int
		}
		SetLog  bool
		Version string // 驱动应该用什么版本的api与服务器对话，最好和服务器版本保持一致以避免怪问题
	}
	saramaConfig *cluster.Config
	consumer     *cluster.Consumer
	output       chan *sarama.ConsumerMessage

	sync.WaitGroup
}

func CreateConsumer(config *config.Config, path string) (c *KafkaConsumer, err error) {
	c = new(KafkaConsumer)
	err = config.GetByScan(path, &c.config)
	if err != nil {
		return nil, err
	}

	cconf := cluster.NewConfig()
	cconf.ClientID = c.config.ClientId
	// 对应配置
	cconf.Consumer.Return.Errors = true
	cconf.Group.Return.Notifications = true

	if c.config.SASL.Enable {
		cconf.Net.SASL.Enable = true
		cconf.Net.SASL.User = c.config.SASL.User
		cconf.Net.SASL.Password = c.config.SASL.Password
	}

	// 配置自动提交, fixme sarama-cluster包还是自动提交，默认用的CommitInterval，Shopify/sarama包已经改为可以配置不用自动提交了。后续需要修改
	cconf.Consumer.Offsets.CommitInterval = time.Duration(c.config.AutoCommit.Interval) * time.Second // 旧版本
	if c.config.AutoCommit.Enable {
		cconf.Consumer.Offsets.AutoCommit.Interval = time.Duration(c.config.AutoCommit.Interval) * time.Second
	}

	if c.config.Version != "" {
		kafkaVersion, err := sarama.ParseKafkaVersion(c.config.Version)
		if err == nil {
			cconf.Version = kafkaVersion
			logger.Debugf("Kafka 版本配置为: %s", kafkaVersion)
		} else {
			logger.Errorf("错误的 Kafka 版本配置: %s", c.config.Version)
			os.Exit(1)
		}
	}

	//cconf.Net.ReadTimeout = 30 * time.Second
	c.saramaConfig = cconf
	c.consumer, err = cluster.NewConsumer(c.config.Brokers, c.config.GroupId, c.config.Topics, c.saramaConfig)
	if err != nil {
		return nil, err
	}
	c.output = make(chan *sarama.ConsumerMessage, 100)

	if c.config.SetLog {
		sarama.Logger = newSaramaLogger()
	}

	return c, nil
}
func (c *KafkaConsumer) Start() {

	go func() {
		c.Add(1)
		c.Worker()
		c.Done()
	}()
}
func (c *KafkaConsumer) Worker() {

	for {
		select {
		case msg, ok := <-c.consumer.Messages():
			if !ok {
				// 收到关闭通知
				close(c.output)
				return
			}
			c.output <- msg

			// fixme : 做成执行完业务才能提交偏移量。但是看依赖包的mloop是自动提交的。
			c.consumer.MarkOffset(msg, "")
			c.consumer.CommitOffsets()
		case n, ok := <-c.consumer.Notifications():
			if ok {
				logger.Infof("%s发出通知: %s", c, n.Type)
				logger.Infof("减持分区 %+v", n.Released)
				logger.Infof("增持分区 %+v", n.Claimed)
				logger.Infof("当前持有 %+v", n.Current)
			}
		case err, ok := <-c.consumer.Errors():
			if !ok {
				return
			}
			logger.Errorf("%s出错: %s", c, err)
		}
	}

}
func (c *KafkaConsumer) Stop() {
	c.consumer.Close()

	c.Wait()
}

func (c *KafkaConsumer) String() string {
	return fmt.Sprintf("消费者(name=%s, groupId=%s, topics=%v)", c.config.Name, c.config.GroupId, c.config.Topics)
}

func (c *KafkaConsumer) Output() chan *sarama.ConsumerMessage {
	return c.output
}

// kafka Client 会定时提交offset，这里提供主动提交接口
func (s *KafkaConsumer) Commit(msgs ...sarama.ConsumerMessage) {
	for _, m := range msgs {
		s.consumer.MarkPartitionOffset(m.Topic, m.Partition, m.Offset, "")
	}
	s.consumer.CommitOffsets()
}

type saramaLogger struct {
	logger *mlog.ServerLogger
}

func (s *saramaLogger) Print(v ...interface{}) {
	s.logger.Info(fmt.Sprint(v...))
}
func (s *saramaLogger) Printf(format string, v ...interface{}) {
	s.logger.Infof(format, v...)
}
func (s *saramaLogger) Println(v ...interface{}) {
	s.logger.Info(fmt.Sprint(v...))
}
func newSaramaLogger() (sl *saramaLogger) {
	sl = &saramaLogger{
		logger: mlog.NewLogger("saramaLogger"),
	}

	return
}
