package logger

import "go.uber.org/zap"

// 为了兼容pitaya的logger
type _LoggerImp struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

func (l *_LoggerImp) Fatal(format ...interface{}) {
	l.sugar.Fatal(format...)
}
func (l *_LoggerImp) Fatalf(format string, args ...interface{}) {
	l.sugar.Fatalf(format, args...)
}
func (l *_LoggerImp) Fatalln(args ...interface{}) {
	l.sugar.Fatal(args...)
}

func (l *_LoggerImp) Debug(args ...interface{}) {
	l.sugar.Debug(args...)
}
func (l *_LoggerImp) Debugf(format string, args ...interface{}) {
	l.sugar.Debugf(format, args...)
}
func (l *_LoggerImp) Debugln(args ...interface{}) {
	l.sugar.Debug(args...)
}

func (l *_LoggerImp) Error(args ...interface{}) {
	l.sugar.Error(args...)
}
func (l *_LoggerImp) Errorf(format string, args ...interface{}) {
	l.sugar.Errorf(format, args...)
}
func (l *_LoggerImp) Errorln(args ...interface{}) {
	l.sugar.Error(args...)
}

func (l *_LoggerImp) Info(args ...interface{}) {
	l.sugar.Info(args...)
}
func (l *_LoggerImp) Infof(format string, args ...interface{}) {
	l.sugar.Infof(format, args...)
}
func (l *_LoggerImp) Infoln(args ...interface{}) {
	l.sugar.Info(args...)
}

func (l *_LoggerImp) Warn(args ...interface{}) {
	l.sugar.Warn(args...)
}
func (l *_LoggerImp) Warnf(format string, args ...interface{}) {
	l.sugar.Warnf(format, args...)
}
func (l *_LoggerImp) Warnln(args ...interface{}) {
	l.sugar.Warn(args...)
}
