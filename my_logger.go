package my_log

////////////////////////////////////////////////////
// Time        : 2016/6/5 21:33
// Author      : wilsonloo21@163.com
// File        : my_logger.go.go
// Software    : PyCharm
// Description : log 机制
////////////////////////////////////////////////////

import (
	"fmt"
)

const (
	LOG_LEVEL_DEBUG = iota
	LOG_LEVEL_WARNING
	LOG_LEVEL_INFO
	LOG_LEVEL_ERROR
	LOG_LEVEL_FATAL
)

var (
	g_log_level = LOG_LEVEL_DEBUG
)

// 设置log级别
func SetLogLevel(lvl int) {
	g_log_level = lvl
}

func Debugf(format string, args ...interface{}) {
	fmt.Printf("[DEBUG]: " + format, args...)
}

func Debugln(args ...interface{})  {
	fmt.Printf("[DEBUG]: ", args...)
	fmt.Printf("\n")
}

func Warningf(format string, args ...interface{}) {
	fmt.Printf("[WARN]: " + format, args...)
}

func Warningln(args ...interface{})  {
	fmt.Printf("[WARN]: ", args...)
	fmt.Printf("\n")
}

func Infof(format string, args ...interface{}) {
	fmt.Printf("[Info]: " + format, args...)
}

func Infoln(args ...interface{})  {
	fmt.Printf("[Info]: ", args...)
	fmt.Printf("\n")
}