package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

var isDebug bool

func Init(mode string) {
	isDebug = mode == "debug"
	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/log.txt", dir)
	f, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	log.SetOutput(f)
	log.SetFlags(0)
	Info(fmt.Sprintf("【%s】start print logs to file: %s", mode, path))
}

// Debug 开发调试时信息的输出，程序运行正常
func Debug(msg ...string) {
	if isDebug {
		print2("D", msg...)
	}
}

// Info 常规信息的输出，程序运行正常
func Info(msg ...string) {
	print2("I", msg...)
}

// Warning 警告信息的输出，重要，需要尽快去查看，但不需要立刻终止程序
func Warning(msg ...string) {
	print2("W", msg...)
}

// Error 发生重大错误，程序无法运行下去，会调用os.Exit()终止程序；
// 对于调用的第三方包，若希望进行异常recover，也在recover后进行调用，以确保打印信息后退出
func Error(msg ...string) {
	print2("E", msg...)
	os.Exit(1)
}

func print2(level string, msg ...string) {
	_, file, line, _ := runtime.Caller(2)
	fs := strings.Split(file, "/")
	text := fmt.Sprintf("[%s] %s %s:%d: %s", level, time.Now().Format("2006/01/02 15:04:05.000000"), fs[len(fs)-1], line, strings.Join(msg, ""))
	log.Println(text)
	fmt.Println(text)
}
