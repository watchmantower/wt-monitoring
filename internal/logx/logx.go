package logx

import (
	"fmt"
	"time"
)

func Info(format string, args ...any) {
	log("INFO", format, args...)
}

func Warn(format string, args ...any) {
	log("WARN", format, args...)
}

func Error(format string, args ...any) {
	log("ERROR", format, args...)
}

func log(level string, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s [%s] %s\n", time.Now().Format(time.RFC3339), level, message)
}
