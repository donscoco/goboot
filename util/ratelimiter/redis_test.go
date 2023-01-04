package ratelimiter

import (
	"github.com/go-redis/redis"
	"log"
	"testing"
	"time"
)

type RedisProxy struct {
	ProxyName string
	Username  string
	Password  string
	Addrs     []string
	Database  int

	isCluster bool

	// todo 根据 redis.ClusterOptions 的配置项添加
	DialTimeout        int
	ReadTimeout        int
	WriteTimeout       int
	MaxRetries         int
	PoolSize           int
	IdleTimeout        int
	IdleCheckFrequency int

	clusterClient *redis.ClusterClient
	sigleClient   *redis.Client
}

func TestCreateLimiter(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		PoolSize: 200,
	})
	client.Set("key1", "val1", 100*time.Second)

	// 每次间隔Every秒，第一次能使用10个，后续每个间隔 Every秒
	// 实际上是每秒产生，也就是说虽然设置500ms,但不是每500ms一个，而是，每秒会有 1s/500ms=2 个
	limiter1, err := CreateLimiter(Every(500*time.Millisecond), 10, "Test", client)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(limiter1.Allow()) // true
	//fmt.Println(limiter1.Allow()) // false

	for {
		if limiter1.Allow() {
			log.Println(time.Now().Second())
		}
	}

}
