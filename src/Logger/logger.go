package mylogger

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func FileLogger(logFolder string) (*zap.Logger, error) {
	// Check if the log folder exists, create it if not
	if _, err := os.Stat(logFolder); os.IsNotExist(err) {
		err := os.Mkdir(logFolder, 0755)
		if err != nil {
			return nil, errors.Wrap(err, "FileLogger failed")
		}
	}
	filename := filepath.Join(logFolder, "log.log")

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	logFile, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

func SetLogger() *zap.Logger {
	currentDir, err := os.Getwd()
	if err != nil {
		errors.Wrap(err, "Getwd failed")
		log.Fatal(err)
	}

	logPath := filepath.Join(currentDir, "/logs")

	logger, err := FileLogger(logPath)
	if err != nil {
		errors.Wrap(err, "SetLogger failed")
		log.Fatal(err)
	}

	return logger
}
