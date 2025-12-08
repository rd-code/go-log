package log

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type fileOperator struct {
	Dir   string
	level Level
	Name  string
	Time  string
}

// 创建日志所在目录
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

// 创建日志文件和链接文件
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

// 生成文件名和链接名
func (f *fileOperator) generateFileAndLinkName() (fileName, linkName string) {
	fileName = fmt.Sprintf("%s_%s_%s.log", f.Name, fileTag[f.level], f.Time)
	linkName = fmt.Sprintf("%s_%s.log", f.Name, fileTag[f.level])
	return
}

// 生成系统使用的文件
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

const TIME_FORMAT = "2006-01-02T15:04:05.000"

const FILE_SUFFIX_TIME_FORMAT = "010215"

// 定义日志等级，有debug,infonotice,warning,error 级别
type Level = int

const (
	DEBUG Level = iota
	INFO
	NOTICE
	WARNING
	ERROR
)

// 日志级别总数
const levelNum = 5

var levelName = [levelNum]string{
	DEBUG:   "D",
	INFO:    "I",
	NOTICE:  "N",
	WARNING: "W",
	ERROR:   "E",
}

var fileTag = [levelNum]string{
	DEBUG:   "debug",
	INFO:    "info",
	NOTICE:  "notice",
	WARNING: "warn",
	ERROR:   "error",
}

// 打印日志配置信息
type Options struct {
	//日志所在目录
	Dir string
	//日志名称
	Name string
	//日志打印级别
	Level Level
	//是否输出到终端
	StdOut bool
}

// 日志打印具体功能实现
type loggingT struct {
	options *Options
	out     [levelNum]*os.File
	//调用层级
	callSkip int
	//当前日志所在时刻
	current string
	//错误信息
	err error
	mu  sync.Mutex
}

type loggingMsg struct {
	level Level
	data  []byte
	//日志时刻
	current string
}

// 设置日志输出流
func (l *loggingT) initOut(current string) (err error) {
	operator := &fileOperator{
		Dir:  l.options.Dir,
		Name: l.options.Name,
		Time: current,
	}
	for i := DEBUG; i <= ERROR; i++ {
		operator.level = i
		var file *os.File
		if file, err = operator.generate(); err != nil {
			return
		}
		l.out[i] = file
	}
	l.current = current
	return
}

// 处理日志时间和文件的对应
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

// 将具体日志打印到文件
func (l *loggingT) write(level Level, data []byte) (err error) {
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
	switch level {
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

// 将日志信息写入到通道
func (l *loggingT) produce(level Level, data []byte, current string) {
	msg := &loggingMsg{
		level:   level,
		data:    data,
		current: current,
	}
	l.handleTime(msg)
	l.write(msg.level, msg.data)
}

// 获取日志header信息
func (l *loggingT) getHeaderField(ctx context.Context, level Level, msg string) ([]*Field, string) {
	_, file, line, _ := runtime.Caller(l.callSkip + 4)
	res := make([]*Field, 0, 4)
	now := time.Now()
	res = append(res, StringField("_level_", levelName[level]),
		StringField("_time_", now.Format(TIME_FORMAT)),
		StringField("_file_", file+":"+strconv.Itoa(line)),
		StringField("_logid_", GetLodId(ctx)),
		StringField("_msg_", msg))
	return res, now.Format(FILE_SUFFIX_TIME_FORMAT)
}

// 生成日志内容信息
func (l *loggingT) generateContent(ctx context.Context, level Level, msg string, fields ...*Field) (*bytes.Buffer, string) {
	headerFields, current := l.getHeaderField(ctx, level, msg)
	headerFields = append(headerFields, fields...)
	buffer := &bytes.Buffer{}
	WriteToBuffer(buffer, headerFields...)
	buffer.WriteByte('\n')
	return buffer, current
}

// 写日志
func (l *loggingT) writeLog(ctx context.Context, level Level, msg string, fields ...*Field) {
	buffer, current := l.generateContent(ctx, level, msg, fields...)
	l.produce(level, buffer.Bytes(), current)
}

// 写debug日志
func (l *loggingT) debug(ctx context.Context, msg string, fields ...*Field) {
	if DEBUG < l.options.Level {
		return
	}
	l.writeLog(ctx, DEBUG, msg, fields...)
}

// 写info日志
func (l *loggingT) info(ctx context.Context, msg string, fields ...*Field) {
	if INFO < l.options.Level {
		return
	}
	l.writeLog(ctx, INFO, msg, fields...)
}

// 写notice日志
func (l *loggingT) notice(ctx context.Context, msg string, fields ...*Field) {
	if NOTICE < l.options.Level {
		return
	}
	l.writeLog(ctx, NOTICE, msg, fields...)
}

// 写warn日志
func (l *loggingT) warn(ctx context.Context, msg string, fields ...*Field) {
	if WARNING < l.options.Level {
		return
	}
	l.writeLog(ctx, WARNING, msg, fields...)
}

// 写error日志
func (l *loggingT) error(ctx context.Context, msg string, fields ...*Field) {
	if ERROR < l.options.Level {
		return
	}
	l.writeLog(ctx, ERROR, msg, fields...)
}

// 增加日志显示调用方跳过的级别
func (l *loggingT) AddCallSkip(callSkip int) {
	l.callSkip = callSkip + callSkip
}

// 写入Debug日志
func (l *loggingT) Debug(ctx context.Context, msg string, fields ...*Field) {
	l.debug(ctx, msg, fields...)
}

// 写入带字符串格式化功能的bebug日志
func (l *loggingT) DebugF(ctx context.Context, format string, args ...interface{}) {
	l.debug(ctx, fmt.Sprintf(format, args...))
}

// 参考Debug
func (l *loggingT) Info(ctx context.Context, msg string, fields ...*Field) {
	l.info(ctx, msg, fields...)
}

// 参考DebugF
func (l *loggingT) InfoF(ctx context.Context, format string, args ...interface{}) {
	l.info(ctx, fmt.Sprintf(format, args...))
}

// 参考Debug
func (l *loggingT) Notice(ctx context.Context, msg string, fields ...*Field) {
	l.notice(ctx, msg, fields...)
}

// 参考DebugF
func (l *loggingT) NoticeF(ctx context.Context, format string, args ...interface{}) {
	l.notice(ctx, fmt.Sprintf(format, args...))
}

// 参考Debug
func (l *loggingT) Warn(ctx context.Context, msg string, fields ...*Field) {
	l.warn(ctx, msg, fields...)
}

// 参考DebugF
func (l *loggingT) WarnF(ctx context.Context, format string, args ...interface{}) {
	l.warn(ctx, fmt.Sprintf(format, args...))
}

// 参考Debug
func (l *loggingT) Error(ctx context.Context, msg string, fields ...*Field) {
	l.error(ctx, msg, fields...)
}

// 参考DebugF
func (l *loggingT) ErrorF(ctx context.Context, format string, args ...interface{}) {
	l.error(ctx, fmt.Sprintf(format, args...))
}

// 创建日志信息
func newLoggingT(options *Options) (res *loggingT, err error) {
	res = &loggingT{options: options, callSkip: 1}
	if err = res.initOut(time.Now().Format(FILE_SUFFIX_TIME_FORMAT)); err != nil {
		return
	}
	return
}

// 日志需要对外提供的功能
type Logger interface {
	//写debug日志
	Debug(ctx context.Context, msg string, fields ...*Field)
	//写格式化Debug日志
	DebugF(ctx context.Context, format string, args ...interface{})
	//写info日志
	Info(ctx context.Context, msg string, fields ...*Field)
	//写格式化info日志
	InfoF(ctx context.Context, format string, args ...interface{})
	//写notice日志
	Notice(ctx context.Context, msg string, fields ...*Field)
	//写格式化notice日志
	NoticeF(ctx context.Context, format string, args ...interface{})
	//写warn日志
	Warn(ctx context.Context, msg string, fields ...*Field)
	//写格式化warn日志
	WarnF(ctx context.Context, format string, args ...interface{})
	//写error日志
	Error(ctx context.Context, msg string, fields ...*Field)
	//写格式化error日志
	ErrorF(ctx context.Context, format string, args ...interface{})
	//设置调用往上跳过级别
	AddCallSkip(callSkip int)
}

var _ Logger = &loggingT{}

// 生成日志模块
func NewLogger(options *Options) (logger Logger, err error) {
	return newLoggingT(options)
}

type LogIdT struct {
}

func WithLogId(ctx context.Context) context.Context {
	return WithLogIdValue(ctx, genLogId())
}

func WithLogIdValue(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, loggingT{}, value)
}

var random = rand.New(rand.NewSource(time.Now().UnixMilli()))

func genLogId() string {
	num := random.Intn(10000000000)
	return fmt.Sprintf("%10d", num)
}

func GetLodId(ctx context.Context) string {
	value := ctx.Value(loggingT{})
	if value == nil {
		return ""
	}
	res, _ := value.(string)
	return res
}
