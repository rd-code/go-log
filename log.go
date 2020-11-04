package log

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
	defaultLogger.SetLevel(2)
	return
}

//写debug日志
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

//写格式化Debug日志
func DebugF(format string, args ...interface{}) {
	defaultLogger.DebugF(format, args...)
}

//参考Debug
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

//参考DebugF
func InfoF(format string, args ...interface{}) {
	defaultLogger.InfoF(format, args...)
}

//参考Debug
func Trace(args ...interface{}) {
	defaultLogger.Trace(args...)
}

//参考DebugF
func TraceF(format string, args ...interface{}) {
	defaultLogger.TraceF(format, args ...)
}

//参考Debug
func Notice(args ...interface{}) {
	defaultLogger.Notice(args...)
}

//参考DebugF
func NoticeF(format string, args ...interface{}) {
	defaultLogger.NoticeF(format, args...)
}

//参考Debug
func War(args ...interface{}) {
	defaultLogger.Warn(args...)
}

//参考DebugF
func WarnF(format string, args ...interface{}) {
	defaultLogger.WarnF(format, args...)
}

//参考Debug
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

//参考DebugF
func ErrorF(format string, args ...interface{}) {
	defaultLogger.ErrorF(format, args...)
}
