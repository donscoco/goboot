package mlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

const (
	LOG_INFO  = 1 << iota // 1
	LOG_WARN              // 2
	LOG_DEBUG             // 4
	LOG_ERROR             // 8
)

const (
	// 不检查更新切换文件
	UpdateModeWithoutDivision int = 0
	// 每小时更新
	UpdateModeHour int = 1
	// 每天更新
	UpdateModeDay int = 2
)

type Logger struct {

	// 创建实例时，传入的文件路径，不可修改不可变的
	Path string
	// 当前带上时间后缀的日志文件路径，根据path生成对应的文件名路径，会根据时间的不同，缓存的文件名不相同
	FileName string
	// 更新模式，暂时只有小时和天的切换文件
	UpdateMode int
	Logger     *log.Logger
	FileCloser io.Closer

	// 参数
	Flags    int
	IsStdout bool
	// 锁
	Lock sync.RWMutex

	LogLevel int // 日志等级
	//Mode       int    // 更新模式类型
	Prefix     string // 前缀
	codePrefix string // code前缀

}

func CreateLogger(path string, logLevel int, updateMode int, prefix string, isStdout bool) (logger *Logger) {
	logger = &Logger{
		Path:       path,
		FileName:   "",
		UpdateMode: updateMode,
		Logger:     nil,
		FileCloser: nil,
		Flags:      0,
		IsStdout:   isStdout,
		LogLevel:   logLevel,
		Prefix:     "{" + prefix + "} ",
	}

	if path != "" {
		// 初始化文件读取
		logger.initFileStream()
		// 初始化更新模式
		logger.initUpdateMode()
	} else {
		// 无文件模式，直接输出到stdout
		logger.initStdout()
	}

	return
}

func (l *Logger) getPath() (path string, err error) {
	path = l.Path
	switch l.UpdateMode {
	case UpdateModeDay:
		day := time.Now().Format("20060102")
		path = path + "." + day
	case UpdateModeHour:
		hour := time.Now().Format("2006010215")
		path = path + "." + hour
	}
	path, err = filepath.Abs(path)
	return
}

func (l *Logger) getTimer() (t *time.Timer) {
	now := time.Now()
	var next time.Time
	switch l.UpdateMode {
	// 按小时更新
	case UpdateModeHour:
		next = now.Add(time.Hour)
		next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location())
	// 按天更新
	case UpdateModeDay:
		next = now.Add(time.Hour * 24)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
	}

	t = time.NewTimer(next.Sub(now))
	return t
}

// log 输出到标准流
func (l *Logger) initStdout() {
	var w io.Writer
	if l.IsStdout {
		w = io.MultiWriter(os.Stdout)
	} else {
		w = io.MultiWriter(os.Stderr)
	}
	l.Logger = log.New(w, "", log.LstdFlags)

	l.FileCloser = os.Stdout
}

// log 输出到文件
func (l *Logger) initFileStream() {
	//1.打开文件
	fileWithSuff, _ := l.getPath()
	f, err := os.OpenFile(fileWithSuff, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	l.FileName = fileWithSuff // 记录一下文件

	//2.将打开的文件描述符作为logger的输出对象
	var w io.Writer
	if l.IsStdout {
		w = io.MultiWriter(os.Stdout, f)
	} else {
		w = f
	}
	l.Logger = log.New(w, "", log.LstdFlags)

	//3.保留文件描述符——为了后续关闭
	l.FileCloser = f
}
func (l *Logger) initUpdateMode() {
	// 0为不更新检查文件切换
	if l.UpdateMode == 0 {
		return
	}
	// 定时检查更新log文件
	go func() {
		var timer *time.Timer
		for {
			// 定时唤醒
			timer = l.getTimer()
			<-timer.C

			// 更换新文件 fix:close 使用close的还有正在打印的协程，所以要加锁
			l.Lock.Lock()
			l.close()
			l.initFileStream()
			l.Lock.Unlock()

		}
	}()
}

// close 关闭文件
func (l *Logger) close() {
	l.FileCloser.Close()
}

func (l *Logger) Log(v ...interface{}) {
	v = append([]interface{}{l.Prefix}, v...)

	l.Lock.RLock()
	l.Logger.Output(2, fmt.Sprint(v...))
	l.Lock.RUnlock()
}

func (l *Logger) Logf(format string, v ...interface{}) {
	format = l.Prefix + format

	l.Lock.RLock()
	l.Logger.Output(2, fmt.Sprintf(format, v...))
	l.Lock.RUnlock()
}
func (l *Logger) LogStack(format string, v ...interface{}) {
	stack := debug.Stack()
	str := fmt.Sprintf(format, v...)
	str = fmt.Sprintf(str, "\n", string(stack))

	l.Lock.RLock()
	l.Logger.Output(2, str)
	l.Lock.RUnlock()
}
