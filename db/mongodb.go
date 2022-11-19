package db

import (
	"context"
	"goboot/config"
	"strconv"
	"time"

	//"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
	https://pkg.go.dev/gopkg.in/mgo.v2
	https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo
*/

const (
	DefaultIdCounter = "sys.IdCounter"
)

type MongoDBProxy struct {
	ProxyName   string
	Username    string
	Password    string
	PasswordSet bool
	Addrs       []string
	Database    string

	MaxPoolSize int
	MinPoolSize int
	ReplicaSet  string
	Mechanism   string
	Timeout     int

	client *mongo.Client
}

var DEFAULT_IDCOUNTER = "sys.IdCounter"

// MongoDb连接对象
// todo fixme
func CreateMongoDBProxy(config *config.Config, path string, manager *DBManager) (err error) {

	for i := 0; i < config.GetInt(path+"/length"); i++ {
		p := new(MongoDBProxy)
		err = config.GetByScan(path+"/"+strconv.Itoa(i), p)
		if err != nil {
			return err
		}

		//clientOpts := &options.ClientOptions{
		//	AppName:                  nil,
		//	Auth:                     nil,
		//	AutoEncryptionOptions:    nil,
		//	ConnectTimeout:           nil,
		//	Compressors:              nil,
		//	Dialer:                   nil,
		//	Direct:                   nil,
		//	DisableOCSPEndpointCheck: nil,
		//	HeartbeatInterval:        nil,
		//	Hosts:                    p.Addrs,
		//	HTTPClient:               nil,
		//	LoadBalanced:             nil,
		//	LocalThreshold:           nil,
		//	MaxConnIdleTime:          nil,
		//	MaxPoolSize:              nil,
		//	MinPoolSize:              nil,
		//	MaxConnecting:            nil,
		//	PoolMonitor:              nil,
		//	Monitor:                  nil,
		//	ServerMonitor:            nil,
		//	ReadConcern:              nil,
		//	ReadPreference:           nil,
		//	Registry:                 nil,
		//	ReplicaSet:               nil,
		//	RetryReads:               nil,
		//	RetryWrites:              nil,
		//	ServerAPIOptions:         nil,
		//	ServerSelectionTimeout:   nil,
		//	SRVMaxHosts:              nil,
		//	SRVServiceName:           nil,
		//	Timeout:                  &t,
		//	TLSConfig:                nil,
		//	WriteConcern:             nil,
		//	ZlibLevel:                nil,
		//	ZstdLevel:                nil,
		//	AuthenticateToAnything:   nil,
		//	Crypt:                    nil,
		//	Deployment:               nil,
		//	SocketTimeout:            nil,
		//}

		clientOpts := options.Client().
			SetAuth(options.Credential{
				//AuthMechanism:           "",
				//AuthMechanismProperties: nil,
				//AuthSource:              "",
				Username:    p.Username,
				Password:    p.Password,
				PasswordSet: p.PasswordSet,
			}).
			SetConnectTimeout(time.Duration(p.Timeout) * time.Second).
			SetHosts(p.Addrs).
			SetMaxPoolSize(uint64(p.MaxPoolSize)).
			SetMinPoolSize(uint64(p.MinPoolSize)).
			SetReplicaSet(p.ReplicaSet)
		client, err := mongo.NewClient(clientOpts)
		if err != nil {
			return err
		}
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			return err
		}

		p.client = client
		manager.MongoDB[p.ProxyName] = p

	}
	return
}
