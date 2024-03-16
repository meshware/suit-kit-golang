package log

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level

const (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
	PanicLevel = zapcore.PanicLevel
	FatalLevel = zapcore.FatalLevel
)

type Logger struct {
	l *zap.Logger
	// https://pkg.go.dev/go.uber.org/zap#example-AtomicLevel
	al *zap.AtomicLevel
}

// InitLog log instance init
func InitLog() {
	viper.SetDefault("log.level", "info")
	level := viper.GetString("log.level")
	logLevel := InfoLevel
	if "debug" == strings.ToLower(level) {
		logLevel = DebugLevel
	}
	if "info" == strings.ToLower(level) {
		logLevel = InfoLevel
	}
	if "error" == strings.ToLower(level) {
		logLevel = ErrorLevel
	}
	if "warn" == strings.ToLower(level) {
		logLevel = WarnLevel
	}
	fmt.Println("Set logLevel = " + level)
	var options = make([]Option, 0)
	logConfig := NewProductionRotateConfig("log")
	appName := viper.GetString("log.appName")
	if len(appName) > 0 {
		logConfig.Filename = appName
	}
	viper.SetDefault("log.development", true)
	development := viper.GetBool("log.development")
	if development {
		options = append(options, Development())
	}
	debugFileName := viper.GetString("log.debugFileName")
	if len(debugFileName) > 0 {
		//modOpts = append(modOpts, setDebugFileName(debugFileName))
	}
	infoFileName := viper.GetString("log.infoFileName")
	if len(infoFileName) > 0 {
		//modOpts = append(modOpts, setInfoFileName(infoFileName))
	}
	errorFileName := viper.GetString("log.errorFileName")
	if len(errorFileName) > 0 {
		//modOpts = append(modOpts, setErrorFileName(errorFileName))
	}
	maxAge := viper.GetInt("log.maxAge")
	if maxAge > 0 {
		logConfig.MaxAge = maxAge
	}
	maxBackups := viper.GetInt("log.maxBackups")
	if maxBackups > 0 {
		logConfig.MaxBackups = maxBackups
	}
	maxSize := viper.GetInt("log.maxSize")
	if maxSize > 0 {
		logConfig.MaxSize = maxSize
	}
	var iow io.Writer
	if !development {
		iow = NewRotateBySize(logConfig)
	}
	std = New(iow, logLevel, options...)
	defer Sync()
}

func New(out io.Writer, level Level, opts ...Option) *Logger {
	var cfg zapcore.EncoderConfig
	// 创建写入同步器
	var fileSyncer zapcore.WriteSyncer
	consoleSyncer := zapcore.AddSync(os.Stdout)
	if out == nil {
		out = os.Stderr
		cfg = zap.NewDevelopmentEncoderConfig()
	} else {
		fileSyncer = zapcore.AddSync(out)
		cfg = zap.NewProductionEncoderConfig()
	}
	al := zap.NewAtomicLevelAt(level)
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder
	// 创建zap核心
	var core zapcore.Core
	if fileSyncer == nil {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(cfg),
			consoleSyncer,
			al,
		)
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(cfg),
				consoleSyncer,
				al,
			),
			zapcore.NewCore(
				zapcore.NewJSONEncoder(cfg),
				fileSyncer,
				al,
			),
		)
	}
	return &Logger{l: zap.New(core, opts...), al: &al}
}

// SetLevel 动态更改日志级别
// 对于使用 NewTee 创建的 Logger 无效，因为 NewTee 本意是根据不同日志级别
// 创建的多个 zap.Core，不应该通过 SetLevel 将多个 zap.Core 日志级别统一
func (l *Logger) SetLevel(level Level) {
	if l.al != nil {
		l.al.SetLevel(level)
	}
}

type Field = zap.Field

func (l *Logger) Debug(msg string, fields ...Field) {
	l.l.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.l.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.l.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.l.Error(msg, fields...)
}

func (l *Logger) Panic(msg string, fields ...Field) {
	l.l.Panic(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	l.l.Fatal(msg, fields...)
}

func (l *Logger) Sync() error {
	return l.l.Sync()
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func timeUnixNano(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.UnixNano() / 1e6)
}

var std = New(os.Stderr, InfoLevel)

func Default() *Logger         { return std }
func ReplaceDefault(l *Logger) { std = l }

func SetLevel(level Level) { std.SetLevel(level) }

func Debug(msg string, fields ...Field) { std.Debug(msg, fields...) }
func Info(msg string, fields ...Field)  { std.Info(msg, fields...) }
func Warn(msg string, fields ...Field)  { std.Warn(msg, fields...) }
func Error(msg string, fields ...Field) { std.Error(msg, fields...) }
func Panic(msg string, fields ...Field) { std.Panic(msg, fields...) }
func Fatal(msg string, fields ...Field) { std.Fatal(msg, fields...) }

func Sync() error { return std.Sync() }
