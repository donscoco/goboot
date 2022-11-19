package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"goboot/config"
	"log"
	"testing"
	"time"
)

func TestCreateMySQLProxy(t *testing.T) {
	config.ConfigFilePath = "../config-demo.json"
	conf := config.NewConfiguration(config.ConfigFilePath)

	dbm := new(DBManager)
	dbm.MySQL = make(map[string]*MySQLProxy)
	err := CreateMySQLProxy(conf, "/core/dbManager/mysql", dbm)
	if err != nil {
		log.Fatal(err)
	}
	p := dbm.MySQL["ironhead-mysql"]
	row, err := p.Session.Query("SELECT * FROM test.iron_student")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(row.Columns())
}

func TestCreateRedisProxy(t *testing.T) {
	config.ConfigFilePath = "../config-demo.json"
	conf := config.NewConfiguration(config.ConfigFilePath)

	dbm := new(DBManager)
	dbm.Redis = make(map[string]*RedisProxy)
	err := CreateRedisProxy(conf, "/core/dbManager/redis", dbm)
	if err != nil {
		log.Fatal(err)
	}
	p := dbm.Redis["ironhead-redis"]
	ok, err := p.sigleClient.SetNX("testkey", "testval", time.Duration(100*time.Second)).Result()
	if err != nil {
	}
	fmt.Println(ok)
	result, err := p.sigleClient.Get("testkey").Result()
	if err != nil {
	}
	fmt.Println(result)
}

// todo
func TestCreateMongoDBProxy(t *testing.T) {
	//config.ConfigFilePath = "../config-demo.json"
	//c := config.NewConfiguration(config.ConfigFilePath)
	//m := new(DBManager)
	//m.MongoDB = make(map[string]*MongoDBProxy)
	//err := CreateMongoDBProxy(c, "/core/dbManager/mongodb", m)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//p, ok := m.MongoDB["ironhead-mongodb"]
	//if !ok {
	//
	//}
	//err = p.client.Ping(context.TODO(), nil)
	//if err != nil {
	//
	//}

	/////
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://192.168.2.128:27017,192.168.2.128:27018,192.168.2.128:27019")

	// 连接到MongoDB
	mgoCli, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// 检查连接
	err = mgoCli.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

}
