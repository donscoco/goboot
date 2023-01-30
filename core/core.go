package core

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/donscoco/goboot/config"
	"github.com/donscoco/goboot/db"
	"github.com/donscoco/goboot/log/mlog"
)

var logger = mlog.NewLogger(`bootcore`)
var GoCore *Core

type Core struct {
	config *config.Config

	DB *db.DBManager

	signal chan os.Signal

	startActs []func() error
	stopActs  []func() error
}

func NewCore() (core *Core) {
	core = new(Core)
	core.config = config.NewConfiguration(config.ConfigFilePath)
	core.signal = make(chan os.Signal, 1)
	return
}

// Boot 运行服务器消息监听循环，等待外部进程消息
func (a *Core) Boot(arg ...func()) {
	// 执行开机操作
	a.start()

	// 等待系统信号
	a.loop()

	// 优雅退出
	a.stop()

	return
}

// 设置开机动作
func (a *Core) OnStart(starts ...func() error) {
	a.startActs = starts
}

// 设置关机动作
func (a *Core) OnStop(stops ...func() error) {
	a.stopActs = stops
}

func (a *Core) start() {

	mlog.InitLoggerByConfig(a.GetConf(), "/core/log")

	logger.Infof("应用开始启动")

	//todo 启动core的组件

	// 执行业务开机动作
	for _, startAct := range a.startActs {
		err := startAct()
		if err != nil {
			logger.Errorf("启动出错：%s", err)
			os.Exit(1)
		}
	}

	logger.Infof(`应用启动完成[PID=%d]`, os.Getpid())
}
func (a *Core) loop() {

	signal.Notify(a.signal, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM) // 监听信号

	for {
		sig := <-a.signal
		logger.Infof("接收到信号: %s", sig)

		switch sig {
		case syscall.SIGUSR1: // 自定义操作，例如设置为debug日志模式
			// todo
		case syscall.SIGUSR2:
			// todo
		default: // sigint 和 sigterm 就退出
			return
		}
	}

}
func (a *Core) stop() {
	logger.Infof("应用开始关闭")

	var err error
	// 执行业务关闭动作
	for _, stopAct := range a.stopActs {
		err = stopAct()
		if err != nil {
			logger.Errorf("关闭出错: %s", err)
		}
	}

	//todo 关闭core的组件
	mlog.Close()

	logger.Infof("应用关闭完成")
}

func (a *Core) GetConf() (c *config.Config) {
	return a.config
}

func InitDBManager() error {

	// 初始化数据库连接
	a := GoCore
	if a == nil {
		err := fmt.Errorf("GoCore 未初始化")
		logger.Error(err)
		return err
	}

	if !GoCore.GetConf().Exist(db.DBConfigPath) {
		err := fmt.Errorf("配置" + db.DBConfigPath + "不存在")
		logger.Error(err)
		return err
	}

	logger.Infof("启动数据库管理器")
	a.DB = new(db.DBManager)

	if a.GetConf().Exist(db.DBConfigPath + "/mysql") {
		a.DB.MySQL = make(map[string]*db.MySQLProxy)
		err := db.CreateMySQLProxy(a.GetConf(), db.DBConfigPath+"/mysql", a.DB)
		if err != nil {
			logger.Error(err)
			return err
		}
		logger.Infof("启动[mysql]数据库管理器")
	}

	if a.GetConf().Exist(db.DBConfigPath + "/redis") {
		a.DB.Redis = make(map[string]*db.RedisProxy)
		err := db.CreateRedisProxy(a.GetConf(), db.DBConfigPath+"/redis", a.DB)
		if err != nil {
			logger.Error(err)
			return err
		}
		logger.Infof("启动[redis]数据库管理器")
	}

	if a.GetConf().Exist(db.DBConfigPath + "/mongodb") {
		a.DB.MongoDB = make(map[string]*db.MongoDBProxy)
		err := db.CreateMongoDBProxy(a.GetConf(), db.DBConfigPath+"/mongodb", a.DB)
		if err != nil {
			logger.Error(err)
			return err
		}
		logger.Infof("启动[mongodb]数据库管理器")
	}

	return nil

}
