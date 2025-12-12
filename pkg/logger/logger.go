package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Global logger instance
	globalLogger zerolog.Logger
)

// Level 日志级别
type Level int8

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled
)

// Config 日志配置
type Config struct {
	Level      Level     // 日志级别
	Output     io.Writer // 输出目标
	TimeFormat string    // 时间格式
	Pretty     bool      // 美化输出（开发模式）
}

// Init 初始化全局日志器
func Init(cfg Config) {
	// 设置日志级别
	zerolog.SetGlobalLevel(zerolog.Level(cfg.Level))

	// 设置时间格式
	if cfg.TimeFormat != "" {
		zerolog.TimeFieldFormat = cfg.TimeFormat
	} else {
		zerolog.TimeFieldFormat = time.RFC3339
	}

	// 创建日志器
	var output io.Writer = cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// 美化输出（开发模式）
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: "2006-01-02 15:04:05",
		}
	}

	globalLogger = zerolog.New(output).With().Timestamp().Caller().Logger()
	log.Logger = globalLogger
}

// WithContext 在 context 中添加日志器
func WithContext(ctx context.Context, logger zerolog.Logger) context.Context {
	return logger.WithContext(ctx)
}

// FromContext 从 context 中获取日志器
func FromContext(ctx context.Context) zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// WithField 添加字段
func WithField(key string, value interface{}) zerolog.Logger {
	return globalLogger.With().Interface(key, value).Logger()
}

// WithFields 添加多个字段
func WithFields(fields map[string]interface{}) zerolog.Logger {
	logger := globalLogger.With()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	return logger.Logger()
}

// Debug 输出 debug 级别日志
func Debug(msg string) {
	globalLogger.Debug().Msg(msg)
}

// Debugf 格式化输出 debug 级别日志
func Debugf(format string, v ...interface{}) {
	globalLogger.Debug().Msgf(format, v...)
}

// DebugWithFields 输出带字段的 debug 日志
func DebugWithFields(msg string, fields map[string]interface{}) {
	logger := globalLogger.Debug()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	logger.Msg(msg)
}

// Info 输出 info 级别日志
func Info(msg string) {
	globalLogger.Info().Msg(msg)
}

// Infof 格式化输出 info 级别日志
func Infof(format string, v ...interface{}) {
	globalLogger.Info().Msgf(format, v...)
}

// InfoWithFields 输出带字段的 info 日志
func InfoWithFields(msg string, fields map[string]interface{}) {
	logger := globalLogger.Info()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	logger.Msg(msg)
}

// Warn 输出 warn 级别日志
func Warn(msg string) {
	globalLogger.Warn().Msg(msg)
}

// Warnf 格式化输出 warn 级别日志
func Warnf(format string, v ...interface{}) {
	globalLogger.Warn().Msgf(format, v...)
}

// WarnWithFields 输出带字段的 warn 日志
func WarnWithFields(msg string, fields map[string]interface{}) {
	logger := globalLogger.Warn()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	logger.Msg(msg)
}

// Error 输出 error 级别日志
func Error(msg string) {
	globalLogger.Error().Msg(msg)
}

// Errorf 格式化输出 error 级别日志
func Errorf(format string, v ...interface{}) {
	globalLogger.Error().Msgf(format, v...)
}

// ErrorWithFields 输出带字段的 error 日志
func ErrorWithFields(msg string, fields map[string]interface{}) {
	logger := globalLogger.Error()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	logger.Msg(msg)
}

// ErrorWithErr 输出带错误的 error 日志
func ErrorWithErr(err error, msg string) {
	globalLogger.Error().Err(err).Msg(msg)
}

// Fatal 输出 fatal 级别日志并退出
func Fatal(msg string) {
	globalLogger.Fatal().Msg(msg)
}

// Fatalf 格式化输出 fatal 级别日志并退出
func Fatalf(format string, v ...interface{}) {
	globalLogger.Fatal().Msgf(format, v...)
}

// FatalWithErr 输出带错误的 fatal 日志并退出
func FatalWithErr(err error, msg string) {
	globalLogger.Fatal().Err(err).Msg(msg)
}

// Panic 输出 panic 级别日志并 panic
func Panic(msg string) {
	globalLogger.Panic().Msg(msg)
}

// Panicf 格式化输出 panic 级别日志并 panic
func Panicf(format string, v ...interface{}) {
	globalLogger.Panic().Msgf(format, v...)
}

// GetLogger 获取全局日志器
func GetLogger() zerolog.Logger {
	return globalLogger
}
