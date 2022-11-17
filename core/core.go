package core

import (
	"goboot/log/mlog"
	"os"
	"os/signal"
	"syscall"
)

var logger = mlog.NewLogger(`bootcore`)
var GoCore *Core

type Core struct {
	signal chan os.Signal

	startActs []func() error
	stopActs  []func() error
}

func NewCore() (core *Core) {
	core = new(Core)
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
	logger.Infof("应用开始启动")

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

	logger.Infof("应用关闭完成")
}
