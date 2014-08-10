package log

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// levels
const (
	debugLevel   = 0
	releaseLevel = 1
	errorLevel   = 2
	fatalLevel   = 3
)

const (
	printDebugLevel   = "[debug  ] "
	printReleaseLevel = "[release] "
	printErrorLevel   = "[error  ] "
	printFatalLevel   = "[fatal  ] "
)

type Logger struct {
	level      int
	baseLogger *log.Logger
}

func New(strLevel string, pathname string) (*Logger, error) {
	// level
	var level int
	switch strings.ToLower(strLevel) {
	case "debug":
		level = debugLevel
	case "release":
		level = releaseLevel
	case "error":
		level = errorLevel
	case "fatal":
		level = fatalLevel
	default:
		return nil, errors.New("unknown level: " + strLevel)
	}

	// logger
	var baseLogger *log.Logger
	if pathname != "" {
		now := time.Now()

		filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second())

		f, err := os.Create(path.Join(pathname, filename))
		if err != nil {
			return nil, err
		}

		baseLogger = log.New(f, "", log.LstdFlags)
	} else {
		baseLogger = log.New(os.Stdout, "", log.LstdFlags)
	}

	// new
	logger := new(Logger)
	logger.level = level
	logger.baseLogger = baseLogger

	return logger, nil
}

func (logger *Logger) doPrintf(level int, printLevel string, format string, a ...interface{}) {
	if level < logger.level {
		return
	}

	format = printLevel + format
	logger.baseLogger.Printf(format, a...)

	if level == fatalLevel {
		os.Exit(1)
	}
}

func (logger *Logger) Debug(format string, a ...interface{}) {
	logger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func (logger *Logger) Release(format string, a ...interface{}) {
	logger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

func (logger *Logger) Error(format string, a ...interface{}) {
	logger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

func (logger *Logger) Fatal(format string, a ...interface{}) {
	logger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

// global logger
var globalLogger *Logger

func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

func Debug(format string, a ...interface{}) {
	globalLogger.Debug(format, a...)
}

func Release(format string, a ...interface{}) {
	globalLogger.Release(format, a...)
}

func Error(format string, a ...interface{}) {
	globalLogger.Error(format, a...)
}

func Fatal(format string, a ...interface{}) {
	globalLogger.Fatal(format, a...)
}
