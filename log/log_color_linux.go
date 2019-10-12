package log

import (
	"fmt"
	"runtime/debug"
)

func (logger *Logger) OutputWithColor(level int, calldepth int, s string) {
	/**
	debugLevel   = 0
	releaseLevel = 1
	errorLevel   = 2
	fatalLevel   = 3
	*/

	str := ""
	switch level {
	case debugLevel: //Debug 绿色
		str = logger.LinuxPatchColor(s, 32)
	case releaseLevel: //release 粉色
		str = logger.LinuxPatchColor(s, 35)
	case errorLevel: //error 黄色
		str = logger.LinuxPatchColor(s, 33)
	case fatalLevel: //fatal 红色
		str = logger.LinuxPatchColor(s, 31)
	default:
		str = logger.LinuxPatchColor(s, 34)
	}

	logger.baseLogger.Output(3, str)

	if level >= errorLevel {
		logger.baseLogger.Output(3, string(debug.Stack()))
	}
}

func (logger *Logger) LinuxPatchColor(str string, fontColor int) string {

	pfTp := 1
	/**
		//  0  终端默认设置
	    //  1  高亮显示
	    //  4  使用下划线
	    //  5  闪烁
	    //  7  反白显示
	    //  8  不可见
	*/
	bgColor := 40

	s := fmt.Sprintf(" %c[%d;%d;%dm%s%c[0m ", 0x1B, pfTp, bgColor, fontColor, str, 0x1B)
	//fmt.Printf("\n %c[1;40;32m%s%c[0m\n\n", 0x1B, "testPrintColor", 0x1B)
	//1, 40 ,32 高亮 背景色 前景色
	return s
}
