package db

import (
	"database/sql"
	"fmt"
	"goboot/config"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLProxy struct {
	ProxyName string
	Username  string
	Password  string
	Addr      string
	Database  string

	Session *sql.DB

	// 连接时设置的参数，参考mysql driver 包的 parseDSNParams 进行添加
	ReadTimeoutSec  int
	WriteTimeoutSec int
	ConnTimeoutSec  int

	// 建立连接后设置的参数
	ConnMaxLifetime int // 设置链接的生命周期
	MaxIdleConns    int // 设置闲置链接数
	MaxOpenConns    int // 设置最大链接数

}

func CreateMySQLProxy(config *config.Config, path string, manager *DBManager) (err error) {
	for i := 0; i < config.GetInt(path+"/length"); i++ {
		// 获取配置路径
		p := new(MySQLProxy)
		err = config.GetByScan(path+"/"+strconv.Itoa(i), p)
		if err != nil {
			//todo
			return err
		}
		// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s)/%s?timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
			p.Username, p.Password, p.Addr, p.Database,
			p.ConnTimeoutSec,
			p.ReadTimeoutSec,
			p.WriteTimeoutSec,
		)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			//todo
			return err
		}

		err = db.Ping()
		if err != nil {
			return err
		}

		// 设置链接的生命周期
		if p.ConnMaxLifetime != 0 {
			db.SetConnMaxLifetime(time.Second * time.Duration(int64(p.ConnMaxLifetime)))
		}
		// 设置闲置链接数
		if p.MaxIdleConns != 0 {
			db.SetMaxIdleConns(p.MaxIdleConns)
		}
		// 设置最大链接数
		if p.MaxOpenConns != 0 {
			db.SetMaxOpenConns(p.MaxOpenConns)
		}

		p.Session = db
		manager.MySQL[p.ProxyName] = p
	}
	return
}
