package log

import "context"

/**
 * DESCRIPTION:
 *
 * @author rd
 * @create 2018-12-08 17:48
 **/

//默认日志实现类
var defaultLogger Logger = nil

//初始化默认日志实现类
func InitDefault(options *Options) (err error) {
	defaultLogger, err = NewLogger(options)
	if err != nil {
		return
	}
	defaultLogger.AddCallSkip(1)
	return
}

func AddCallSkip(callSkip int) {
	defaultLogger.AddCallSkip(callSkip)
}

//写debug日志
func Debug(ctx context.Context, msg string, fields ...*Field) {
	defaultLogger.Debug(ctx, msg, fields...)
}

//写格式化Debug日志
func DebugF(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.DebugF(ctx, format, args...)
}

//参考Debug
func Info(ctx context.Context, msg string, fields ...*Field) {
	defaultLogger.Info(ctx, msg, fields...)
}

//参考DebugF
func InfoF(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.InfoF(ctx, format, args...)
}

//参考Debug
func Notice(ctx context.Context, msg string, fields ...*Field) {
	defaultLogger.Notice(ctx, msg, fields...)
}

//参考DebugF
func NoticeF(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.NoticeF(ctx, format, args...)
}

//参考Debug
func Warn(ctx context.Context, msg string, fields ...*Field) {
	defaultLogger.Warn(ctx, msg, fields...)
}

//参考DebugF
func WarnF(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.WarnF(ctx, format, args...)
}

//参考Debug
func Error(ctx context.Context, msg string, fields ...*Field) {
	defaultLogger.Error(ctx, msg, fields...)
}

//参考DebugF
func ErrorF(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.ErrorF(ctx, format, args...)
}
