package mredis

import (
	"github.com/go-redis/redis"
	"github.com/donscoco/goboot/config"
	"github.com/donscoco/goboot/log/mlog"
	"time"
)

var logger = mlog.NewLogger("mq-redis")

type RedisPubSub struct {
	Config struct {
		Topic string

		Addr     string
		Username string
		Password string
		Database int

		// todo 根据 redis.ClusterOptions 的配置项添加
		DialTimeout        int
		ReadTimeout        int
		WriteTimeout       int
		MaxRetries         int
		PoolSize           int
		IdleTimeout        int
		IdleCheckFrequency int
	}
	running bool
	handler map[string]func(topic string, param []byte) error

	client *redis.Client
	pubsub map[string]*redis.PubSub
}

// todo 将db改成server ,统一管理所有涉及网络IO的服务组件，然后这里创建函数通过配置文件指定使用哪一个server
func CreateRedisMQ(config *config.Config, path string) (rps *RedisPubSub, err error) {
	r := new(RedisPubSub)

	err = config.GetByScan(path, &r.Config)
	if err != nil {
		return nil, err
	}
	c := &r.Config

	opt := &redis.Options{
		Addr: c.Addr,
		//Password:     p.Password,
		DB:           c.Database,
		DialTimeout:  time.Second * time.Duration(c.DialTimeout),
		ReadTimeout:  time.Second * time.Duration(c.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(c.WriteTimeout),
		MaxRetries:   c.MaxRetries,
	}
	if len(c.Password) > 0 {
		opt.Password = c.Password
	}
	r.client = redis.NewClient(opt)

	cmd := r.client.Ping()
	if cmd.Val() != "PONG" {
		return nil, cmd.Err()
	}

	r.pubsub = make(map[string]*redis.PubSub)
	r.handler = make(map[string]func(topic string, param []byte) error)

	return r, nil

}

func (q *RedisPubSub) Publish(topic string, msg string) error {
	cmd := q.client.Publish(topic, msg)
	r, err := cmd.Result()
	if err != nil {
		return err
	}
	logger.Debugf("Publish success num %d", r)
	return nil

}
func (q *RedisPubSub) Subscribe(topic string, callback func(topic string, param []byte) error) error {
	pubsub := q.client.Subscribe(topic)
	q.pubsub[topic] = pubsub
	q.handler[topic] = callback
	return nil
}

func (q *RedisPubSub) Start() {
	q.running = true

	for topic, pubsub := range q.pubsub {
		go func(t string, ps *redis.PubSub) {
			for q.running {
				msg, err := ps.ReceiveMessage()
				if err != nil && q.running {
					logger.Error(err)
				}
				if err != nil && q.running == false {
					return
				}
				// fixme
				q.handler[t](t, []byte(msg.Payload))
			}
		}(topic, pubsub)
	}

}
func (q *RedisPubSub) Stop(p interface{}) error {
	q.running = false

	for _, p := range q.pubsub {
		p.Close()

	}
	if q.client != nil {
		q.client.Close()
	}

	return nil
}
