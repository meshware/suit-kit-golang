package log

import (
	"os"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	Info("failed to fetch URL", String("url", "https://jianghushinian.cn/"))
	Warn("Warn msg", Int("attempt", 3))
	Error("Error msg", Duration("backoff", time.Second))

	// 修改日志级别
	SetLevel(ErrorLevel)
	Info("Info msg")
	Warn("Warn msg")
	Error("Error msg")

	// 替换默认 Logger
	file, _ := os.OpenFile("custom.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	logger := New(file, InfoLevel)
	ReplaceDefault(logger)
	Info("Info msg in replace default logger after")
}
