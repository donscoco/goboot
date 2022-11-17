package core

import (
	"goboot/log/mlog"
	"testing"
)

// go test -v -run TestCore ./core
func TestCore(t *testing.T) {
	//os.Setenv(mlog.LogPathEnvName, "/tmp/logpath")
	//os.Setenv(mlog.LogLevelEnvName, "INFO")
	//os.Setenv(mlog.LogModeEnvName, "hour")
	//mlog.InitLogger()
	// 先设置环境变量
	// export GOBOOT_LOG_PATH=/tmp/logpath && export GOBOOT_LOG_LEVEL=INFO && export GOBOOT_LOG_MODE=hour

	GoCore = NewCore()
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
