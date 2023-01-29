package lock

import (
	"github.com/go-redis/redis"
	"github.com/donscoco/goboot/config"
	"log"
	"time"
)

// 初始化创建redis

var DefaultRedisLocker *RedisLocker

type RedisLocker struct {
	Username string
	Password string
	Addrs    []string
	Database int

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

	Expire   int
	LockName string
}

func Init(conf *config.Config, path string) {
	var err error
	DefaultRedisLocker, err = CreateRedisLocker(conf, path, "defaultlock", 160)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateRedisLocker(config *config.Config, path string, LockName string, Expire int) (l *RedisLocker, err error) {

	// 读取配置
	l = new(RedisLocker)
	err = config.GetByScan(path, l)
	if err != nil {
		return nil, err
	}

	if len(l.Addrs) > 1 {
		l.isCluster = true

		// 创建client
		opt := &redis.ClusterOptions{
			Addrs:        l.Addrs,
			Password:     l.Password,
			DialTimeout:  time.Second * time.Duration(l.DialTimeout),
			ReadTimeout:  time.Second * time.Duration(l.ReadTimeout),
			WriteTimeout: time.Second * time.Duration(l.WriteTimeout),
			MaxRetries:   l.MaxRetries,
		}
		if len(l.Password) > 0 {
			opt.Password = l.Password
		}
		l.clusterClient = redis.NewClusterClient(opt)
		cmd := l.clusterClient.Ping()
		if cmd.Val() != "PONG" {
			return nil, cmd.Err()
		}

	} else if len(l.Addrs) == 1 { // 单节点
		l.isCluster = false
		opt := &redis.Options{
			Addr: l.Addrs[0],
			//Password:     p.Password,
			DB:           l.Database,
			DialTimeout:  time.Second * time.Duration(l.DialTimeout),
			ReadTimeout:  time.Second * time.Duration(l.ReadTimeout),
			WriteTimeout: time.Second * time.Duration(l.WriteTimeout),
			MaxRetries:   l.MaxRetries,
		}
		if len(l.Password) > 0 {
			opt.Password = l.Password
		}
		l.sigleClient = redis.NewClient(opt)

		cmd := l.sigleClient.Ping()
		if cmd.Val() != "PONG" {
			return nil, cmd.Err()
		}

	} else {
		// empty addr
	}

	l.Expire = Expire
	l.LockName = LockName

	return
}

// trylock
func (l *RedisLocker) TryLock() (isLocked bool, err error) {
	currentTime := time.Now().Unix()
	var bcmd *redis.BoolCmd
	if l.isCluster {
		bcmd = l.clusterClient.SetNX(l.LockName, currentTime, time.Duration(l.Expire)*time.Second)
	} else {
		bcmd = l.sigleClient.SetNX(l.LockName, currentTime, time.Duration(l.Expire)*time.Second)

	}
	return bcmd.Val(), bcmd.Err()
}

// unlock
func (l *RedisLocker) UnLock() (err error) {
	if l.isCluster {
		return l.clusterClient.Del(l.LockName).Err()
	} else {
		return l.sigleClient.Del(l.LockName).Err()
	}

}
