package db

import (
	//redis "gopkg.in/redis.v5"
	"github.com/go-redis/redis"
	"github.com/donscoco/goboot/config"
	"strconv"
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

func CreateRedisProxy(config *config.Config, path string, manager *DBManager) (err error) {
	for i := 0; i < config.GetInt(path+"/length"); i++ {
		// 读取配置
		p := new(RedisProxy)
		err = config.GetByScan(path+"/"+strconv.Itoa(i), p)
		if err != nil {
			return err
		}

		if len(p.Addrs) > 1 {
			p.isCluster = true

			// 创建client
			opt := &redis.ClusterOptions{
				Addrs:        p.Addrs,
				Password:     p.Password,
				DialTimeout:  time.Second * time.Duration(p.DialTimeout),
				ReadTimeout:  time.Second * time.Duration(p.ReadTimeout),
				WriteTimeout: time.Second * time.Duration(p.WriteTimeout),
				MaxRetries:   p.MaxRetries,
			}
			if len(p.Password) > 0 {
				opt.Password = p.Password
			}
			p.clusterClient = redis.NewClusterClient(opt)
			cmd := p.clusterClient.Ping()
			if cmd.Val() != "PONG" {
				return cmd.Err()
			}

		} else if len(p.Addrs) == 1 { // 单节点
			p.isCluster = false
			opt := &redis.Options{
				Addr: p.Addrs[0],
				//Password:     p.Password,
				DB:           p.Database,
				DialTimeout:  time.Second * time.Duration(p.DialTimeout),
				ReadTimeout:  time.Second * time.Duration(p.ReadTimeout),
				WriteTimeout: time.Second * time.Duration(p.WriteTimeout),
				MaxRetries:   p.MaxRetries,
			}
			if len(p.Password) > 0 {
				opt.Password = p.Password
			}
			p.sigleClient = redis.NewClient(opt)

			cmd := p.sigleClient.Ping()
			if cmd.Val() != "PONG" {
				return cmd.Err()
			}

		} else {
			// empty addr
		}
		manager.Redis[p.ProxyName] = p

	}
	return
}
