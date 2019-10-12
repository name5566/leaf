package log

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
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
	baseFile   *os.File
	HourFlag   string
	pathname   string
	flag       int
}

func New(strLevel string, pathname string, flag int) (*Logger, error) {
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
	var baseFile *os.File
	HF := ""
	if pathname != "" {
		now := time.Now()
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second())
		full := path.Join(dir+"/"+pathname, filename)
		file, err := os.Create(full)
		if err != nil {
			return nil, err
		}

		HF = fmt.Sprintf("%d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour())

		baseLogger = log.New(file, "", flag)
		baseFile = file
	} else {
		baseLogger = log.New(os.Stdout, "", flag)
	}

	// new
	logger := new(Logger)
	logger.level = level
	logger.baseLogger = baseLogger
	logger.baseFile = baseFile
	logger.HourFlag = HF
	logger.pathname = pathname
	logger.flag = flag

	return logger, nil
}
func (logger *Logger) UpdateFileName() {

	//非日期文件不变更
	if logger.HourFlag == "" {
		return
	}

	//没换时间则跳过
	now := time.Now()
	HF := fmt.Sprintf("%d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour())
	if HF == logger.HourFlag {
		finfo, _ := logger.baseFile.Stat()
		if finfo.Size() < 1024*1024*1024 {
			return
		}

	}

	//换时间的直接更换文件
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second())

	full := path.Join(dir+"/"+logger.pathname, filename)
	file, err := os.Create(full)
	if err != nil {
		return
	}

	baseLogger := log.New(file, "", logger.flag)
	logger.baseLogger = baseLogger
	logger.baseFile = file
	logger.HourFlag = HF

	//logger.HourFlag

}

// It's dangerous to call the method on logging
func (logger *Logger) Close() {
	if logger.baseFile != nil {
		logger.baseFile.Close()
	}

	logger.baseLogger = nil
	logger.baseFile = nil
}

func (logger *Logger) doPrintf(level int, printLevel string, format string, a ...interface{}) {
	if level < logger.level {
		return
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}
	//更新文件名
	logger.UpdateFileName()
	//isWin
	_, file, line, _ := runtime.Caller(2)
	format = printLevel + "[" + file + "][" + strconv.Itoa(line) + "]" + format
	logger.OutputWithColor(level, 3, fmt.Sprintf(format, a...))

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
	_, file, line, _ := runtime.Caller(1)
	logger.doPrintf(errorLevel, printErrorLevel, "%v %v", file, line)
	debug.PrintStack()
}

func (logger *Logger) Fatal(format string, a ...interface{}) {
	logger.doPrintf(fatalLevel, printFatalLevel, format, a...)
	_, file, line, _ := runtime.Caller(1)
	logger.doPrintf(errorLevel, printErrorLevel, "%v %v", file, line)
	debug.PrintStack()
}

var gLogger, _ = New("debug", "", log.LstdFlags)

// It's dangerous to call the method on logging
func Export(logger *Logger) {
	if logger != nil {
		gLogger = logger
	}
}

func Debug(format string, a ...interface{}) {
	gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func Release(format string, a ...interface{}) {
	gLogger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

func Error(format string, a ...interface{}) {
	gLogger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

func Fatal(format string, a ...interface{}) {
	gLogger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

func DebugJson(format string, a ...interface{}) {
	for k, v := range a {
		if btv, err := json.Marshal(v); err == nil {
			a[k] = string(btv)
		}
	}
	gLogger.doPrintf(releaseLevel, printDebugLevel, format, a...)
}

func ReleaseJson(format string, a ...interface{}) {
	for k, v := range a {
		if btv, err := json.Marshal(v); err == nil {
			a[k] = string(btv)
		}
	}
	gLogger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

func ErrorJson(format string, a ...interface{}) {
	for k, v := range a {
		if btv, err := json.Marshal(v); err == nil {
			a[k] = string(btv)
		}
	}
	gLogger.doPrintf(releaseLevel, printErrorLevel, format, a...)
}

func FatalJson(format string, a ...interface{}) {
	for k, v := range a {
		if btv, err := json.Marshal(v); err == nil {
			a[k] = string(btv)
		}
	}
	gLogger.doPrintf(releaseLevel, printFatalLevel, format, a...)
}

func Close() {
	gLogger.Close()
}
func isWin() bool {

	return runtime.GOOS == "windows"
}
