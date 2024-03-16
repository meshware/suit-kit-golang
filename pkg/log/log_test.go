package log

import (
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	InitLog()
	Info("failed to fetch URL", String("url", "https://www.jd.com"))
	Warn("Warn msg", Int("attempt", 3))
	Error("Error msg", Duration("backoff", time.Second))

	// 修改日志级别
	SetLevel(ErrorLevel)
	Info("Info msg")
	Warn("Warn msg")
	Error("Error msg")
}
