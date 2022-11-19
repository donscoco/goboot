package mlog

import (
	"os"
	"strconv"
	"sync"
	"testing"
)

//go test -v -run TestLog ./log/mlog
func TestLog(t *testing.T) {

	var wg sync.WaitGroup

	os.Setenv(LogPathEnvName, "/tmp/logpath")
	os.Setenv(LogLevelEnvName, "INFO")
	os.Setenv(LogModeEnvName, "hour")

	InitLoggerByEnv()

	Info("test info")
	Debug("test debug")
	Warn("test warn")
	Error("test error")

	for i := 0; i < 10; i++ {
		testServer := NewLogger("server-" + strconv.Itoa(i))
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			testServer.Infof("test info %d", num)
			testServer.Debugf("test info %d", num)
			testServer.Warnf("test info %d", num)
			testServer.Errorf("test info %d", num)
			if num == 5 {
				Close()
			}
		}(i)
	}

	wg.Wait()

}
