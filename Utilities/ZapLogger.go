package Utilities

import (
	"fmt"
	"log"
	"net/url"
	"runtime"

	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

type lumberjackSink struct {
	*lumberjack.Logger
}
type zapLogger struct {
	ZLog *zap.Logger
}

func (lumberjackSink) Sync() error {
	return nil
}

func ZapLog() *zapLogger {
	var logPath string
	os := runtime.GOOS
	if os == "windows" {
		logPath = "./Logs/"
	} else {
		logPath = getenv("BASE_LOG_PATH", "/var/log/tradingview/demo/")
	}
	logFile := logPath + "logs.txt"
	ll := lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, //MB
		MaxBackups: 0,
		MaxAge:     20, //days
		Compress:   true,
	}
	err := zap.RegisterSink("lumberjack", func(*url.URL) (zap.Sink, error) {
		return lumberjackSink{
			Logger: &ll,
		}, nil
	})
	if err != nil {
		return nil
	}
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{fmt.Sprintf("lumberjack:%s", logFile)}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("Status: Failed to initialize zap logger: %v", err)
	}
	logger.Info("Zap Log initiated")
	return &zapLogger{ZLog: logger}
}

var zLgr = ZapLog()
var ZapLogger = zLgr.ZLog
