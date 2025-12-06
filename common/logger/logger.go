package logger

import (
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"staking-interaction/common/config"
	"strings"
	"sync"
	"time"
)

var (
	log  *logrus.Logger
	once sync.Once
)

// InitLogger 初始化日志（项目启动时调用一次即可）
func InitLogger() {
	once.Do(func() {
		log = logrus.New()
		log.SetOutput(os.Stdout)

		conf := config.Get()
		logConfig := conf.LogConfig

		logPath := logConfig.LogPath
		logFileName := logConfig.LogFile
		fileName := path.Join(logPath, logFileName)
		if err := os.MkdirAll(logPath, 0755); err != nil {
			log.Errorf("logger: create directory error:%v", err)
			return
		}

		if conf.AppConfig.Debug {
			log.SetLevel(logrus.DebugLevel)
		} else {
			log.SetLevel(logrus.InfoLevel)
		}

		// 日志格式
		var formatter logrus.Formatter
		if logConfig.IsJsonFormat {
			formatter = &logrus.JSONFormatter{TimestampFormat: config.TIME_FORMAT}
		} else {
			formatter = &logrus.TextFormatter{FullTimestamp: true, TimestampFormat: config.TIME_FORMAT}
		}
		log.SetFormatter(formatter)
		writeFile(logPath, fileName, time.Duration(logConfig.MaxAge), formatter)
	})
}

// GetLogger 提供全局 logger
func GetLogger() *logrus.Logger {
	if log == nil {
		InitLogger() // 默认配置
	}
	return log
}

// writeFile 设置日志切割和钩子
func writeFile(logPath string, filename string, maxAge time.Duration, formatter logrus.Formatter) {
	// Info、Debug、Warn 级别写 info 日志
	infoWriter, err := rotatelogs.New(
		strings.Replace(filename, ".logger", "_info_", -1)+".%Y%m%d.logger",
		rotatelogs.WithLinkName(path.Join(logPath, "info.logger")),
		rotatelogs.WithMaxAge(maxAge*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Errorf("logger: info rotatelogs error:%v", err)
		return
	}

	// Error、Fatal、Panic 级别写 error 日志
	errWriter, err := rotatelogs.New(
		strings.Replace(filename, ".logger", "_err_", -1)+".%Y%m%d.logger",
		rotatelogs.WithLinkName(path.Join(logPath, "error.logger")),
		rotatelogs.WithMaxAge(maxAge*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Errorf("logger: error rotatelogs error:%v", err)
		return
	}

	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  infoWriter,
		logrus.FatalLevel: infoWriter,
		logrus.DebugLevel: infoWriter,
		logrus.WarnLevel:  errWriter,
		logrus.ErrorLevel: errWriter,
		logrus.PanicLevel: errWriter,
	}

	//把 lfshook 钩子添加到 logrus，将日志按照设定级别写入自动切割的文件
	log.AddHook(lfshook.NewHook(writeMap, formatter))
}
