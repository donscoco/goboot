package db

import "github.com/donscoco/goboot/log/mlog"

var logger = mlog.NewLogger("db")

type DBManager struct {
	MySQL   map[string]*MySQLProxy
	Redis   map[string]*RedisProxy
	MongoDB map[string]*MongoDBProxy
}
