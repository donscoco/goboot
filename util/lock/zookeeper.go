package lock

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/donscoco/goboot/config"
	"strconv"
	"strings"
	"time"
)

var DefaultZKLocker *ZKLocker

type zkcfg struct {
	Addrs   []string
	Timeout int
	Path    string
}

func InitZK(conf *config.Config, path string) (err error) {

	cfg := zkcfg{}
	err = conf.GetByScan(path, &cfg)
	if err != nil {
		return err
	}

	conn, _, err := zk.Connect(cfg.Addrs, time.Duration(cfg.Timeout)*time.Second)
	if err != nil {
		return err
	}

	DefaultZKLocker, err = CreateZKLocker(conn, cfg.Path)
	if err != nil {
		return err
	}
	return
}

type ZKLocker struct {
	Conn   *zk.Conn
	path   string
	prefix string

	lockPath string
	seq      int
}

func CreateZKLocker(c *zk.Conn, rootpath string) (locker *ZKLocker, err error) {
	locker = new(ZKLocker)
	locker.Conn = c
	locker.path = rootpath
	locker.prefix = fmt.Sprintf("%s/lock-", locker.path)

	isExists, _, err := locker.Conn.Exists(locker.path)
	if err != nil {
		return nil, err
	}

	if isExists {
		return
	} else { // 创建父节点
		_, err := locker.Conn.Create(locker.path, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return nil, err
		}
	}

	return
}

func (l *ZKLocker) Lock() (err error) {
	// 创建临时有序节点
	// 取出所有子节点
	// 判断最小的， //如果是最小的节点，则表示获取到锁
	// 最小表示获得锁，返回
	// 不是最小的，注册监听
	//此处如果有延时，上一个节点 在此刻被删除，自己最小却无法实现监听
	// 需要再次判断下上一个节点还存不存在
	// 存在，没有获得锁，等待获得锁通知
	// 不存在，说明上一个已经退出了。那么我就是最小的了。直接获得锁，返回，ps：这里有可能最小的没有退出，但是我的上一个节点已经退出了。所以这里最好还是再次获得所有节点，确认下自己是不是最小的。

	path, err := l.Conn.CreateProtectedEphemeralSequential(l.prefix, nil, zk.WorldACL(zk.PermAll))
	if err != nil {
		return
	}

	seq, err := parseSeq(path)
	if err != nil {
		return err
	}

	for {
		children, _, err := l.Conn.Children(l.path)
		if err != nil {
			return err
		}

		lowestSeq := seq
		prevSeq := -1
		prevSeqPath := ""
		for _, p := range children { // 遍历目标：1。找到最小的，2。找到比自己的集合中最大的那个
			s, err := parseSeq(p)
			if err != nil {
				return err
			}
			if s < lowestSeq {
				lowestSeq = s
			}
			if s < seq && s > prevSeq {
				prevSeq = s
				prevSeqPath = p
			}
		}
		if seq == lowestSeq { // 没有找到比自己还小的，说明自己是队首，直接获得锁
			break
		}
		// 没有获得锁，watch等待通知获得锁，为了避免羊群效应，watch自己前面的那个人
		_, _, ch, err := l.Conn.GetW(l.path + "/" + prevSeqPath)
		if err != nil && err != zk.ErrNoNode {
			return err
		} else if err != nil && err == zk.ErrNoNode { // 这里可能出现要watch那个节点退出了。所以要再次看下所有孩子节点的最小
			continue
		}

		ev := <-ch
		if ev.Err != nil {
			return ev.Err
		}
	}

	l.seq = seq
	l.lockPath = path
	return
}
func (l *ZKLocker) Unlock() (err error) {
	if l.lockPath == "" {
		return zk.ErrNotLocked
	}
	if err := l.Conn.Delete(l.lockPath, -1); err != nil {
		return err
	}
	l.lockPath = ""
	l.seq = 0
	return nil
}
func parseSeq(path string) (int, error) {
	parts := strings.Split(path, "-")
	return strconv.Atoi(parts[len(parts)-1])
}
