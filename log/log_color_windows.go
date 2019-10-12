package log

import (
	"runtime/debug"
	"syscall"
)

func (logger *Logger) OutputWithColor(level int, calldepth int, s string) {
	/**
	debugLevel   = 0
	releaseLevel = 1
	errorLevel   = 2
	fatalLevel   = 3
	*/

	switch level {
	case debugLevel: //Debug 绿色
		logger.winPatchColor(calldepth, s, 2|8)
	case releaseLevel: //release 粉色
		logger.winPatchColor(calldepth, s, 5|8)
	case errorLevel: //error 黄色
		logger.winPatchColor(calldepth, s, 6|8)
	case fatalLevel: //fatal 红色
		logger.winPatchColor(calldepth, s, 4|8)
	default: //未知 蓝色
		logger.winPatchColor(calldepth, s, 3|8)
	}
	if level >= errorLevel {
		logger.baseLogger.Output(3, string(debug.Stack()))
	}

}

func (logger *Logger) winPatchColor(calldepth int, s string, color int) {

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("SetConsoleTextAttribute")
	handle, _, _ := proc.Call(uintptr(syscall.Stdout), uintptr(color))
	logger.baseLogger.Output(calldepth, s)
	handle, _, _ = proc.Call(uintptr(syscall.Stdout), uintptr(7))
	CloseHandle := kernel32.NewProc("CloseHandle")
	CloseHandle.Call(handle)

}
