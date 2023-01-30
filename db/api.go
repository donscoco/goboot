package db

import (
	"fmt"
	"github.com/donscoco/goboot/log/mlog"
)

var logger = mlog.NewLogger("db")

const DBConfigPath = "/core/dbManager"

type DBManager struct {
	MySQL   map[string]*MySQLProxy
	Redis   map[string]*RedisProxy
	MongoDB map[string]*MongoDBProxy
}

func (m *DBManager) GetMySQL(name string) (proxy *MySQLProxy, err error) {
	db := m.MySQL[name]
	if db == nil {
		err = fmt.Errorf("找不到 mysql 节点[%s]", name)

		logger.Errorf(err.Error())

		return nil, err
	}
	return db, nil
}

func (m *DBManager) GetRedis(name string) (proxy *RedisProxy, err error) {
	db := m.Redis[name]
	if db == nil {
		err = fmt.Errorf("找不到 redis 节点[%s]", name)

		logger.Errorf(err.Error())

		return nil, err
	}
	return db, nil
}

func (m *DBManager) GetMongo(name string) (proxy *MongoDBProxy, err error) {
	db := m.MongoDB[name]
	if db == nil {
		err = fmt.Errorf("找不到 mongodb 节点[%s]", name)

		logger.Errorf(err.Error())

		return nil, err
	}
	return db, nil
}
