package mlog

import (
	"log"
	"os"
	"path/filepath"

	"github.com/donscoco/goboot/config"
)

// 全局的实例
var InfoLogger *Logger
var InfoLogFile string
var ErrorLogger *Logger
var ErrorLogFile string
var WarnLogger *Logger
var WarnLogFile string
var DebugLogger *Logger

// 环境配置信息
var __logPath string = "./var/log/"
var __level string
var __updateMode string

// ConfigPathEnvName config directory env name
const LogPathEnvName = `GOBOOT_LOG_PATH`
const LogLevelEnvName = `GOBOOT_LOG_LEVEL`
const LogModeEnvName = `GOBOOT_LOG_MODE`

var Level int
var Mode int

// 使用配置文件作为log的参数
func InitLoggerByConfig(conf *config.Config, path string) {
	__logPath = conf.GetString(path + "/logPath")
	__level = conf.GetString(path + "/logLevel")
	__updateMode = conf.GetString(path + "/logMode")

	InitLogger()
}

// 使用环境变量作为log的参数
func InitLoggerByEnv() {
	__logPath = os.Getenv(LogPathEnvName)
	__level = os.Getenv(LogLevelEnvName)
	__updateMode = os.Getenv(LogModeEnvName)

	InitLogger()
}

// InitLogger 初始化日志对象
func InitLogger() {

	initLevel()
	initMode()
	initPath()

	// level为debug模式下，重置文件日志路径，不需要写入到文件中，直接输出到stdout
	if (Level & LOG_DEBUG) != 0 {
		InfoLogFile = ""
		ErrorLogFile = ""
		WarnLogFile = ""
	}

	// 常规日志输出
	InfoLogger = CreateLogger(InfoLogFile, Level, Mode, "INFO", true)
	// 错误输出
	ErrorLogger = CreateLogger(ErrorLogFile, Level, Mode, "ERROR", false)
	// 警告输出
	WarnLogger = CreateLogger(WarnLogFile, Level, Mode, "WARN", true)
	// 调试输出
	DebugLogger = CreateLogger("", Level, Mode, "DEBUG", true)

}
func initLevel() {
	switch __level {
	case "DEBUG":
		Level = LOG_DEBUG | LOG_INFO | LOG_WARN | LOG_ERROR
	case "INFO":
		Level = LOG_INFO | LOG_WARN | LOG_ERROR
	case "WARN":
		Level = LOG_WARN | LOG_ERROR
	case "ERROR":
		Level = LOG_ERROR
	case "PRODUCT":
		Level = LOG_WARN | LOG_ERROR
	default:
		//log.Fatalln("undefined env log level")
		// 默认debug
		log.Println("undefined env log level, use debug level")
		Level = LOG_DEBUG | LOG_INFO | LOG_WARN | LOG_ERROR
	}
	log.Printf("init log level by %s \n", __level)
}
func initMode() {
	switch __updateMode {
	case "day":
		Mode = UpdateModeDay
	case "hour":
		Mode = UpdateModeHour
	default:
		log.Println("undefined env log mode, use UpdateModeWithoutDivision")
		Mode = UpdateModeWithoutDivision
	}
	log.Printf("init log update mode by %s \n", __updateMode)
}
func initPath() {
	// 检查目录
	err := os.MkdirAll(__logPath, 0777)
	if err != nil {
		log.Fatal("mkdir log path error:", err)
	}
	// 检查文件目录是否正常打开
	_, err = os.Stat(__logPath)
	if err != nil {
		log.Fatal("init log path error:", err)
		return
	}

	InfoLogFile, _ = filepath.Abs(filepath.Join(__logPath, "app.log"))
	ErrorLogFile, _ = filepath.Abs(filepath.Join(__logPath, "error.log"))
	WarnLogFile, _ = filepath.Abs(filepath.Join(__logPath, "warn.log"))

	log.Println("log file base:", __logPath)
}

func Debugf(format string, arg ...interface{}) {
	if (Level & LOG_DEBUG) > 0 {
		DebugLogger.Logf(format, arg...)
	}
}
func Debug(arg ...interface{}) {
	if (Level & LOG_DEBUG) > 0 {
		DebugLogger.Log(arg...)
	}
}
func Infof(format string, arg ...interface{}) {
	if (Level & LOG_INFO) > 0 {
		InfoLogger.Logf(format, arg...)
	}
}
func Info(arg ...interface{}) {
	if (Level & LOG_INFO) > 0 {
		InfoLogger.Log(arg...)
	}
}
func Warnf(format string, arg ...interface{}) {
	if (Level & LOG_WARN) > 0 {
		WarnLogger.Logf(format, arg...)
	}
}
func Warn(arg ...interface{}) {
	if (Level & LOG_WARN) > 0 {
		WarnLogger.Log(arg...)
	}
}
func Errorf(format string, arg ...interface{}) {
	if (Level & LOG_ERROR) > 0 {
		ErrorLogger.Logf(format, arg...)
	}
}
func Error(arg ...interface{}) {
	if (Level & LOG_ERROR) > 0 {
		ErrorLogger.Log(arg...)
	}
}

func Close() {
	InfoLogger.close()
	DebugLogger.close()
	WarnLogger.close()
	ErrorLogger.close()
}

// 用于给每个组件创建独立的server
type ServerLogger struct {
	Name string
}

func NewLogger(name string) *ServerLogger {
	return &ServerLogger{Name: `[` + name + `] `}
}
func (s ServerLogger) Debug(arg ...interface{}) {
	arg = append([]interface{}{s.Name}, arg...)
	DebugLogger.Log(arg...)
}
func (s ServerLogger) Debugf(format string, arg ...interface{}) {
	DebugLogger.Logf(s.Name+format, arg...)
}
func (s ServerLogger) Info(arg ...interface{}) {
	arg = append([]interface{}{s.Name}, arg...)
	InfoLogger.Log(arg...)
}
func (s ServerLogger) Infof(format string, arg ...interface{}) {
	InfoLogger.Logf(s.Name+format, arg...)
}
func (s ServerLogger) Warn(arg ...interface{}) {
	arg = append([]interface{}{s.Name}, arg...)
	WarnLogger.Log(arg...)
}
func (s ServerLogger) Warnf(format string, arg ...interface{}) {
	WarnLogger.Logf(s.Name+format, arg...)
}
func (s ServerLogger) Error(arg ...interface{}) {
	arg = append([]interface{}{s.Name}, arg...)
	ErrorLogger.Log(arg...)
}
func (s ServerLogger) Errorf(format string, arg ...interface{}) {
	ErrorLogger.Logf(s.Name+format, arg...)
}
