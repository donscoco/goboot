package core

import (
	"syscall"
	"testing"
	"time"

	"goboot/config"
	"goboot/log/mlog"
)

// go test -v -run TestCore ./core
func TestCore(t *testing.T) {

	config.ConfigFilePath = "../config-demo.json"
	GoCore = NewCore()

	go func() { // ci 下可能会卡住，2秒后自动模拟退出信号
		time.Sleep(2 * time.Second)
		GoCore.signal <- syscall.SIGINT
	}()

	GoCore.OnStart(startFunc1, startFunc2)
	GoCore.OnStop(stopFunc2, stopFunc1)
	GoCore.Boot()
}

var logger1 = mlog.NewLogger("server1")

func startFunc1() error {
	logger1.Info("init server1")
	return nil
}
func stopFunc1() error {
	logger1.Info("close server1")
	return nil
}

var logger2 = mlog.NewLogger("server2")

func startFunc2() error {
	logger2.Info("init server2")
	return nil
}
func stopFunc2() error {
	logger2.Info("close server2")
	return nil
}
