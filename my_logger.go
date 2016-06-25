package my_log

////////////////////////////////////////////////////
// Time        : 2016/6/5 21:33
// Author      : wilsonloo21@163.com
// File        : my_logger.go.go
// Software    : PyCharm
// Description : log 机制
////////////////////////////////////////////////////

import (
	"io"
	"fmt"
	"os"
	"time"
	"container/list"
)

const (
	LOG_LEVEL_ALL = iota
	LOG_LEVEL_DEBUG
	LOG_LEVEL_WARNING
	LOG_LEVEL_INFO
	LOG_LEVEL_ERROR
	LOG_LEVEL_FATAL

	LOG_CHANNEL_SIZE = 1000 // log channel 大小
	FLUSH_LOG_COUNT = 5 // log落地个数
	FLUSH_INTERVAL = 2 // log落地时间间隔（单位：秒）

	SIGNAL_EXIT = 0xFFFFffff // 退出信号
	SIGNAL_FLUSH_NOW = 0xEEEEeeee // 立即落地
)

type Logger struct {

	log_level                 int         // log 级别
	log_filename              string      // 文件名
	caption                   string      // 标题
	flush_log_threshold_count int         // log落地的条目个数

	log_file		  *os.File    // 实际文件

	flushing_logs             *list.List  // 即将落地的log列表
	flush_timer               *time.Timer

	logs_chan                 chan string // log写入的channel
	signal                    chan int    // 信号（主要用于控制log落地协程）
}

// 创建新的 Logger 对象
// @param log_filename 文件名
// @param caption 标题（用以区分不同过的log实例）
func NewLogger(log_filename string, caption string) *Logger{
	logger := new(Logger)
	logger.SetLogLevel(LOG_LEVEL_ALL)
	logger.log_filename = log_filename
	logger.caption = caption
	logger.flush_log_threshold_count = FLUSH_LOG_COUNT

	logger.flush_timer = time.NewTimer(FLUSH_INTERVAL * time.Second)
	logger.flushing_logs = list.New()
	logger.logs_chan = make(chan string, LOG_CHANNEL_SIZE)

	logger.signal = make(chan int, 1)

	if checkFileIsExist(logger.log_filename) {  //如果文件存在
  		f, err := os.OpenFile(logger.log_filename, os.O_APPEND, 0666)  //打开文件
		if err != nil {
			fmt.Printf("failed to open log file %s error: %s\n", logger.log_filename, err.Error())
		} else {
			logger.log_file = f
		}
	 } else {
	  	f, err := os.Create(logger.log_filename)  //创建文件
		if err != nil {
			fmt.Printf("failed to open log file %s error: %s\n", logger.log_filename, err.Error())
		} else {
			logger.log_file = f
		}
	 }

	// 开启落地协程
	go func(){
		for {
			select {
			case val := <-logger.signal:
				if val == SIGNAL_EXIT {
					// 退出信号，需要结束当前协程
					return
				} else if val == SIGNAL_FLUSH_NOW {
					// 立即落地
					logger.do_flush()
				}

			case <-logger.flush_timer.C:
				// 立即落地
				logger.do_flush()

			case flush_log_text := <- logger.logs_chan:
				// 写入 log flushing 列表
				logger.flushing_logs.PushBack(flush_log_text)
			}
		}
	}()

	return logger
}

// 回收
func FreeLogger(log *Logger)  {
	log.signal <- SIGNAL_FLUSH_NOW
	log.signal <- SIGNAL_EXIT

	if log.log_file != nil {
		log.log_file.Close()
		log.log_file = nil
	}
}

// 设置log级别
func (this *Logger)SetLogLevel(lvl int) {
	this.log_level = lvl
}

// 设置log落地的条目个数
func (this *Logger)SetFlushLogCount(flush_count int) {
	this.flush_log_threshold_count = flush_count
}

func (this *Logger)do_flush() {
	if this.log_file == nil {
		return
	}

	for this.flushing_logs.Len() > 0 {
		io.WriteString(this.log_file, this.flushing_logs.Front().Value.(string))
		this.flushing_logs.Remove(this.flushing_logs.Front())
	}
}

// 实际输出log
func (this *Logger)echo(text string) {
	// 输入到终端
	fmt.Println(text)

	// 写入 落地写队列
	this.logs_chan <- text

	// 检测是否达到落地阀值
	if this.flushing_logs.Len() >= this.flush_log_threshold_count {
		this.signal <- SIGNAL_FLUSH_NOW
	}
}

func (this *Logger)echo_fmt(log_level_text string, format string, args ...interface{}) {
	output := fmt.Sprintf("[%s - %s]: " + format, this.caption, log_level_text, args)
	this.echo(output)
}

func (this *Logger)echo_ln(log_level_text string, args ...interface{})  {
	output := fmt.Sprintf("[%s - %s]: ", this.caption, log_level_text)
	for _, v := range(args) {
		output = output + " " + v.(string)
	}
	this.echo(output)
}

// 调试输出，用于开发
func (this *Logger)Debugf(format string, args ...interface{}) {
	this.echo_fmt("DEBUG", format, args...)
}

// 调试输出，用于开发
func (this *Logger)Debugln(args ...interface{})  {
	this.echo_ln("DEBUG", args...)
}

// 警告，用于提示可能存在的问题，但是不会影响功能的运行
func (this *Logger)Warningf(format string, args ...interface{}) {
	this.echo_fmt("WARN", format, args...)
}

// 警告，用于提示可能存在的问题，但是不会影响功能的运行
func (this *Logger)Warningln(args ...interface{})  {
	this.echo_ln("WARN", args...)
}

// 信息输出，用于正常流程的通知、tip、提示，一般都会打开
func (this *Logger)Infof(format string, args ...interface{}) {
	this.echo_fmt("INFO", format, args...)
}

// 信息输出，用于正常流程的通知、tip、提示，一般都会打开
func (this *Logger)Infoln(args ...interface{})  {
	this.echo_ln("INFO", args...)
}

// 错误输出：可能使该应用失败，需要进行维护
func (this *Logger)Errorf(format string, args ...interface{}) {
	this.echo_fmt("ERROR", format, args...)
}

// 错误输出：可能使该应用失败，需要进行维护
func (this *Logger)Errorln(args ...interface{})  {
	this.echo_ln("ERROR", args...)
}

// 失败输出：该应用失败，需要进行维护，程序可能导致宕机，或数据异常
func (this *Logger)Fatalf(format string, args ...interface{}) {
	this.echo_fmt("FATAL", format, args...)
}

// 失败输出：该应用失败，需要进行维护，程序可能导致宕机，或数据异常
func (this *Logger)Fatalln(args ...interface{})  {
	this.echo_ln("FATAL", args...)
}

// 立即flush
func (this *Logger)Flush() {
	this.signal <- SIGNAL_FLUSH_NOW
}

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) (bool) {
	 var exist = true;
	 if _, err := os.Stat(filename); os.IsNotExist(err) {
	  	exist = false;
	 }

	 return exist;
}