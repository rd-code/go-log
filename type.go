package log

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type fileOperator struct {
	Dir      string
	Severity Severity
	Name     string
	Time     string
}

//创建日志所在目录
func (f *fileOperator) createDir() (err error) {
	file, err := os.Open(f.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(f.Dir, os.ModePerm)
		}
		return err
	}
	var info os.FileInfo
	if info, err = file.Stat(); err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	return DirectoryIsFile
}

//创建日志文件和链接文件
func (f *fileOperator) createFileAndLink(fileName, linkName string) (file *os.File, err error) {
	//创建文件
	if file, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
		//创建失败且不是因为文件存在失败
		if !os.IsExist(err) {
			return
		}
	}

	//移除之前的链接文件
	if err = os.Remove(linkName); err != nil {
		//移除文件失败且不是因为文件不存在失败
		if os.IsNotExist(err) {
			goto label
		}
		file.Close()
		return
	}

label:
	//创建文件链接
	if err = os.Link(fileName, linkName); err != nil {
		file.Close()
		return
	}
	return
}

//生成文件名和链接名
func (f *fileOperator) generateFileAndLinkName() (fileName, linkName string) {
	fileName = fmt.Sprintf("%s_%s_%s.log", f.Name, fileTag[f.Severity], f.Time)
	linkName = fmt.Sprintf("%s_%s.log", f.Name, fileTag[f.Severity])
	return
}

//生成系统使用的文件
func (f *fileOperator) generate() (file *os.File, err error) {
	if err = f.createDir(); err != nil {
		return
	}
	fileName, linkName := f.generateFileAndLinkName()

	fileName = filepath.Join(f.Dir, fileName)
	linkName = filepath.Join(f.Dir, linkName)
	if file, err = f.createFileAndLink(fileName, linkName); err != nil {
		return
	}
	return
}

const TIME_FORMAT = "2006-01-02T15:04:05"

const FILE_SUFFIX_TIME_FORMAT = "010215"

//定义日志等级，有debug,info,trace,notice,warning,error 级别
type Severity int

const (
	DEBUG Severity = iota
	INFO
	TRACE
	NOTICE
	WARNING
	ERROR
)

//日志级别总数
const serverityNum = 6

var severityName = [serverityNum]string{
	DEBUG:   "D",
	INFO:    "I",
	TRACE:   "T",
	NOTICE:  "N",
	WARNING: "W",
	ERROR:   "E",
}

var fileTag = [serverityNum]string{
	DEBUG:   "debug",
	INFO:    "info",
	TRACE:   "trace",
	NOTICE:  "notice",
	WARNING: "warn",
	ERROR:   "error",
}

//打印日志配置信息
type Options struct {
	//日志所在目录
	Dir string
	//日志名称
	Name string
	//日志打印级别
	Severity Severity
	//是否输出到终端
	StdOut bool
}

//日志打印具体功能实现
type loggingT struct {
	options *Options
	out     [serverityNum]*os.File
	//调用层级
	level int
	//是否设置了level
	setLevel bool
	//当前日志所在时刻
	current string
	//错误信息
	err error
	mu  sync.Mutex
}

type loggingMsg struct {
	severity Severity
	data     []byte
	//日志时刻
	current string
}

//设置日志输出流
func (l *loggingT) initOut(current string) (err error) {
	operator := &fileOperator{
		Dir:  l.options.Dir,
		Name: l.options.Name,
		Time: current,
	}
	for i := DEBUG; i <= ERROR; i++ {
		operator.Severity = i
		var file *os.File
		if file, err = operator.generate(); err != nil {
			return
		}
		l.out[i] = file
	}
	l.current = current
	return
}

//处理日志时间和文件的对应
func (l *loggingT) handleTime(msg *loggingMsg) {
	if l.current == msg.current {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := DEBUG; i <= ERROR; i++ {
		l.out[i].Close()
	}

	if l.err = l.initOut(msg.current); l.err != nil {
		fmt.Println("init out failed", l.err)
	}

}

//将具体日志打印到文件
func (l *loggingT) write(severity Severity, data []byte) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	//如果需要将信息输出到终端，则使用终端进行打印
	if l.options.StdOut {
		if _, err = os.Stdout.Write(data); err != nil {
			return
		}
	}
	if l.err != nil {
		return
	}

	//高等级日志一定会在低等级日志文件中出现
	switch severity {
	case ERROR:
		if _, err = l.out[ERROR].Write(data); err != nil {
			return
		}
		fallthrough
	case WARNING:
		if _, err = l.out[WARNING].Write(data); err != nil {
			return
		}
		fallthrough
	case NOTICE:
		if _, err = l.out[NOTICE].Write(data); err != nil {
			return
		}
		fallthrough
	case TRACE:
		if _, err = l.out[TRACE].Write(data); err != nil {
			return
		}
		fallthrough
	case INFO:
		if _, err = l.out[INFO].Write(data); err != nil {
			return
		}
		fallthrough
	case DEBUG:
		if _, err = l.out[DEBUG].Write(data); err != nil {
			return
		}

	default:
		err = UnknownLevelError
		return
	}
	return
}

//将日志信息写入到通道
func (l *loggingT) produce(severity Severity, data []byte, current string) {
	msg := &loggingMsg{
		severity: severity,
		data:     data,
		current:  current,
	}
	l.handleTime(msg)
	l.write(msg.severity, msg.data)

}

//获取日志header信息
func (l *loggingT) getHeader(severity Severity, level int) (*bytes.Buffer, string) {
	_, file, line, _ := runtime.Caller(level + 4)
	now := time.Now()
	buffer := &bytes.Buffer{}
	buffer.WriteString(severityName[severity])
	buffer.WriteString(" ")
	buffer.WriteString(now.Format(TIME_FORMAT))
	buffer.WriteString(" ")
	buffer.WriteString(file)
	buffer.WriteString(":")
	buffer.WriteString(strconv.Itoa(line))
	buffer.WriteString(" ->] ")
	return buffer, now.Format(FILE_SUFFIX_TIME_FORMAT)
}

//生成日志内容信息
func (l *loggingT) generateContent(severity Severity, level int, args ...interface{}) (*bytes.Buffer, string) {
	buffer, current := l.getHeader(severity, level)
	buffer.WriteString(fmt.Sprintln(args...))
	return buffer, current
}

//写日志
func (l *loggingT) writeLog(severity Severity, level int, args ...interface{}) {
	buffer, current := l.generateContent(severity, level, args...)
	l.produce(severity, buffer.Bytes(), current)
}

//写debug日志
func (l *loggingT) debug(level int, args ...interface{}) {
	if DEBUG < l.options.Severity {
		return
	}
	l.writeLog(DEBUG, level, args...)
}

//写info日志
func (l *loggingT) info(level int, args ...interface{}) {
	if INFO < l.options.Severity {
		return
	}
	l.writeLog(INFO, level, args...)
}

//写trace日志
func (l *loggingT) trace(level int, args ...interface{}) {
	if TRACE < l.options.Severity {
		return
	}
	l.writeLog(TRACE, level, args...)
}

//写notice日志
func (l *loggingT) notice(level int, args ...interface{}) {
	if NOTICE < l.options.Severity {
		return
	}
	l.writeLog(NOTICE, level, args...)
}

//写warn日志
func (l *loggingT) warn(level int, args ...interface{}) {
	if WARNING < l.options.Severity {
		return
	}
	l.writeLog(WARNING, level, args...)
}

//写error日志
func (l *loggingT) error(level int, args ...interface{}) {
	if ERROR < l.options.Severity {
		return
	}
	l.writeLog(ERROR, level, args...)
}

//设置日志显示调用方跳过的级别
func (l *loggingT) SetLevel(level int) {
	l.setLevel = true
	l.level = level
}

//获取日志跳过层级
func (l *loggingT) getLevel() int {
	if l.setLevel {
		return l.level
	}
	return 1
}

//写入Debug日志
func (l *loggingT) Debug(args ...interface{}) {
	l.debug(l.getLevel(), args...)
}

//写入带字符串格式化功能的bebug日志
func (l *loggingT) DebugF(format string, args ...interface{}) {
	l.debug(l.getLevel(), fmt.Sprintf(format, args...))
}

//参考Debug
func (l *loggingT) Info(args ...interface{}) {
	l.info(l.getLevel(), args...)
}

//参考DebugF
func (l *loggingT) InfoF(format string, args ...interface{}) {
	l.info(l.getLevel(), fmt.Sprintf(format, args...))
}

//参考Debug
func (l *loggingT) Trace(args ...interface{}) {
	l.trace(l.getLevel(), args...)
}

//参考DebugF
func (l *loggingT) TraceF(format string, args ...interface{}) {
	l.trace(l.getLevel(), fmt.Sprintf(format, args...))
}

//参考Debug
func (l *loggingT) Notice(args ...interface{}) {
	l.notice(l.getLevel(), args...)
}

//参考DebugF
func (l *loggingT) NoticeF(format string, args ...interface{}) {
	l.notice(l.getLevel(), fmt.Sprintf(format, args...))
}

//参考Debug
func (l *loggingT) Warn(args ...interface{}) {
	l.warn(l.getLevel(), args...)
}

//参考DebugF
func (l *loggingT) WarnF(format string, args ...interface{}) {
	l.warn(l.getLevel(), fmt.Sprintf(format, args...))
}

//参考Debug
func (l *loggingT) Error(args ...interface{}) {
	l.error(l.getLevel(), args...)
}

//参考DebugF
func (l *loggingT) ErrorF(format string, args ...interface{}) {
	l.error(l.getLevel(), fmt.Sprintf(format, args...))
}

//创建日志信息
func newLoggingT(options *Options) (res *loggingT, err error) {
	res = &loggingT{options: options}
	if err = res.initOut(time.Now().Format(FILE_SUFFIX_TIME_FORMAT)); err != nil {
		return
	}
	return
}

//日志需要对外提供的功能
type Logger interface {
	//写debug日志
	Debug(args ...interface{})
	//写格式化Debug日志
	DebugF(format string, args ...interface{})
	//写info日志
	Info(args ...interface{})
	//写格式化info日志
	InfoF(format string, args ...interface{})
	//写trace日志
	Trace(args ...interface{})
	//写格式化trace日志
	TraceF(format string, args ...interface{})
	//写notice日志
	Notice(args ...interface{})
	//写格式化notice日志
	NoticeF(format string, args ...interface{})
	//写warn日志
	Warn(args ...interface{})
	//写格式化warn日志
	WarnF(format string, args ...interface{})
	//写error日志
	Error(args ...interface{})
	//写格式化error日志
	ErrorF(format string, args ...interface{})
	//设置调用往上跳过级别
	SetLevel(level int)
}

var _ Logger = &loggingT{}

//生成日志模块
func NewLogger(options *Options) (logger Logger, err error) {
	return newLoggingT(options)
}
