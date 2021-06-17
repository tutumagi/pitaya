package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type _LoggingFormat int8

const (
	_JSONFormat _LoggingFormat = iota
	_StackdriverFormat
)

// Debugf logger
func Debugf(format string, args ...interface{}) {
	if l == nil {
		fmt.Printf(fmt.Sprintf("%s\n", format), args...)
		return
	}
	l.sugar.Debugf(format, args...)
}

// Infof logger
func Infof(format string, args ...interface{}) {
	if l == nil {
		fmt.Printf(fmt.Sprintf("%s\n", format), args...)
		return
	}
	l.sugar.Infof(format, args...)
}

// Warnf logger
func Warnf(format string, args ...interface{}) {
	if l == nil {
		// debug.PrintStack()
		fmt.Printf(fmt.Sprintf("%s\n", format), args...)
		return
	}
	l.sugar.Warnf(format, args...)
}

// Errorf logger
func Errorf(format string, args ...interface{}) {
	if l == nil {
		debug.PrintStack()
		fmt.Printf(fmt.Sprintf("%s\n", format), args...)
		return
	}
	l.sugar.Errorf(format, args...)
}

// Panicf logger, log message then Panic
func Panicf(format string, args ...interface{}) {
	if l == nil {
		debug.PrintStack()
		fmt.Printf(fmt.Sprintf("%s\n", format), args...)
		return
	}
	l.sugar.Panicf(format, args...)
}

// Debug logger
func Debug(msg string, fields ...zapcore.Field) {
	if l == nil {
		fmt.Println(msg)
		return
	}
	l.logger.Debug(msg, fields...)
}

// Info logger
func Info(msg string, fields ...zapcore.Field) {
	if l == nil {
		fmt.Println(msg)
		return
	}
	l.logger.Info(msg, fields...)
}

// Warn logger
func Warn(msg string, fields ...zapcore.Field) {
	if l == nil {
		fmt.Println(msg, fields)
		return
	}
	l.logger.Warn(msg, fields...)
}

// Error logger
func Error(msg string, fields ...zapcore.Field) {
	if l == nil {
		fmt.Println(msg, fields)
		return
	}
	l.logger.Error(msg, fields...)
}

// Fatal logger, log message then call os.Exit(-1).
func Fatal(msg string, fields ...zapcore.Field) {
	if l == nil {
		fmt.Println(msg)
		return
	}
	l.logger.Fatal(msg, fields...)
}

// Panic logger, log message then Panic
func Panic(msg string, fields ...zapcore.Field) {
	if l == nil {
		fmt.Println(msg)
		return
	}
	l.logger.Panic(msg, fields...)
}

var l *_LoggerImp

// Init logger initialize
func Init(serverType string, config *viper.Viper) {
	l = &_LoggerImp{}
	l.logger = newLogger(serverType, config)
	l.sugar = l.logger.Sugar()

	// logger, _ := zap.NewDevelopment()
	// sugar = logger.Sugar()

	// sugar.Infof("setup logger")

	// logger.SetLogger(l)
	l.Info("initialize logger")
}

// NewLogger  config
func newLogger(serverType string, config *viper.Viper) *zap.Logger {

	level := config.GetString("logger.level")
	fileDir := config.GetString("logger.dir")
	rotation := config.GetBool("logger.rotation")
	stdout := config.GetBool("logger.stdout")

	// {servertype}.{datetime}.log
	// dateTime := time.Now().Format("_20060102150405-0700")
	file := strings.Join([]string{fileDir, serverType, ".log"}, "")

	zapLevel := zapcore.InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		fmt.Println("Logger level invalid, must be one of: DEBUG, INFO, WARN, or ERROR")
	}

	format := _JSONFormat
	// switch strings.ToLower(config.Format) {
	// case "":
	// 	fallthrough
	// case "json":
	// 	format = _JSONFormat
	// case "stackdriver":
	// 	format = _StackdriverFormat
	// default:
	// 	fmt.Println("Logger mode format invalid, must be one of :'', 'json', or 'stackdriver'")
	// }

	consoleLogger := newJSONLogger(os.Stdout, zapLevel, format)
	var fileLogger *zap.Logger
	if rotation {
		fileLogger = newRotatingJSONFileLogger(config, consoleLogger, file, zapLevel, format)
	} else {
		fileLogger = newJSONFileLogger(consoleLogger, file, zapLevel, format)
	}

	if fileLogger != nil {
		multilLogger := newMultiLogger(consoleLogger, fileLogger)

		if stdout {
			zap.RedirectStdLog(multilLogger)
			return multilLogger
		}
		zap.RedirectStdLog(fileLogger)
		return fileLogger
	}

	zap.RedirectStdLog(consoleLogger)

	return consoleLogger
}

func newJSONFileLogger(consoleLogger *zap.Logger, fileName string, level zapcore.Level, format _LoggingFormat) *zap.Logger {
	if len(fileName) == 0 {
		return nil
	}

	output, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		consoleLogger.Fatal("Could not create log file", zap.Error(err))
		return nil
	}

	return newJSONLogger(output, level, format)
}

func newRotatingJSONFileLogger(config *viper.Viper, consoleLogger *zap.Logger, fileName string, level zapcore.Level, format _LoggingFormat) *zap.Logger {
	if len(fileName) == 0 {
		consoleLogger.Fatal("Rotating log file is enabled but log file name is empty")
		return nil
	}

	logDir := filepath.Dir(fileName)
	if curPath, err := os.Getwd(); err == nil {
		consoleLogger.Info("cur path", zap.String("dir", curPath))
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				consoleLogger.Fatal("Could not create log directory", zap.Error(err))
				return nil
			}
		}
	} else {
		consoleLogger.Fatal("Could not reach cur directory")
		return nil
	}

	jsonEncoder := newJSONEncoder(format)

	writeSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    config.GetInt("logger.maxsize"),
		MaxAge:     config.GetInt("logger.maxage"),
		MaxBackups: config.GetInt("logger.maxbackups"),
		LocalTime:  config.GetBool("logger.localtime"),
		Compress:   config.GetBool("logger.compress"),
	})

	core := zapcore.NewCore(jsonEncoder, writeSyncer, level)
	options := []zap.Option{zap.AddStacktrace(zap.ErrorLevel)}
	return zap.New(core, options...)
}

func newMultiLogger(loggers ...*zap.Logger) *zap.Logger {
	cores := make([]zapcore.Core, 0, len(loggers))
	for _, logger := range loggers {
		cores = append(cores, logger.Core())
	}
	teeCore := zapcore.NewTee(cores...)
	options := []zap.Option{zap.AddStacktrace(zap.ErrorLevel), zap.AddCaller(), zap.AddCallerSkip(1)}
	return zap.New(teeCore, options...)
}

func newJSONLogger(output *os.File, level zapcore.Level, format _LoggingFormat) *zap.Logger {
	jsonEncoder := newJSONEncoder(format)

	core := zapcore.NewCore(jsonEncoder, zapcore.Lock(output), level)
	options := []zap.Option{zap.AddStacktrace(zap.ErrorLevel), zap.AddCaller(), zap.AddCallerSkip(1)}
	return zap.New(core, options...)
}

// Create a new JSON log encoder with the correct settings.
func newJSONEncoder(format _LoggingFormat) zapcore.Encoder {
	if format == _StackdriverFormat {
		return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "severity",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    stackdriverLevelEncoder,
			EncodeTime:     stackdriverTimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		})
	}

	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func stackdriverTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%d%s", t.Unix(), t.Format(".000000000")))
}

func stackdriverLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString("debug")
	case zapcore.InfoLevel:
		enc.AppendString("info")
	case zapcore.WarnLevel:
		enc.AppendString("warning")
	case zapcore.ErrorLevel:
		enc.AppendString("error")
	case zapcore.DPanicLevel:
		enc.AppendString("critical")
	case zapcore.PanicLevel:
		enc.AppendString("critical")
	case zapcore.FatalLevel:
		enc.AppendString("critical")
	default:
		enc.AppendString(fmt.Sprintf("Level(%d)", l))
	}
}
