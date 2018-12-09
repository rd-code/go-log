package log

import (
    "bytes"
    "flag"
    "fmt"
    "io"
    "os"
    "runtime"
    "sync"
    "time"
)

/**
 * DESCRIPTION:
 *
 * @author rd
 * @create 2018-12-08 17:48
 **/

type severity = int

const (
    infoLog severity = iota
    warningLog
    errLog
)
const numSeverity = 3

const severityChar = "IWE"

var severityName = []string{
    infoLog:    "INFO",
    warningLog: "WARNING",
    errLog:     "ERROR",
}

type buffer struct {
    bytes.Buffer
    lines []byte
    next  *buffer
}

func (b *buffer) store(value int) {
    if b.lines != nil {
        b.lines = b.lines[:0]
    }
    for value != 0 {
        b.lines = append(b.lines, byte(value%10))
        value /= 10
    }
}

type writerSync interface {
    io.Writer
    Sync() error
}

type loggingT struct {
    mu         sync.Mutex
    freeListMu sync.Mutex
    freeList   *buffer

    out [numSeverity]writerSync
}

func (l *loggingT) putBuffer(b *buffer) {
    l.freeListMu.Lock()
    defer l.freeListMu.Unlock()
    b.next = l.freeList
    l.freeList = b
}

func (l *loggingT) getBuffer() (*buffer) {
    l.freeListMu.Lock()
    defer l.freeListMu.Unlock()
    b := l.freeList
    if b != nil {
        l.freeList = b.next
        b.next = nil
    } else {
        b = new(buffer)
    }
    b.Reset()
    return b
}

func (l *loggingT) write(s severity, data *buffer) error {
    l.mu.Lock()
    defer l.mu.Unlock()
    switch s {
    case errLog:
        if _, err := l.out[errLog].Write(data.Bytes()); err != nil {
            return err
        }
        fallthrough
    case warningLog:
        if _, err := l.out[warningLog].Write(data.Bytes()); err != nil {
            return err
        }
        fallthrough
    case infoLog:
        if _, err := l.out[infoLog].Write(data.Bytes()); err != nil {
            return err
        }
    }
    return nil
}

func (l *loggingT) formatHeader(s severity, depth int) *buffer {
    b := l.getBuffer()
    _, file, line, _ := runtime.Caller(3 + depth)
    b.WriteByte(severityChar[s])
    b.WriteByte(' ')
    b.WriteString(file)
    b.WriteByte(':')

    b.store(line)

    for i := len(b.lines) - 1; i >= 0; i-- {
        b.WriteByte(b.lines[i] + '0')
    }

    b.WriteByte(' ')

    b.WriteString(time.Now().Format("2006-01-02T15:04:05"))
    b.WriteByte(' ')
    b.WriteByte(']')
    b.WriteByte(' ')
    return b

}

func (l *loggingT) print(s severity, depth int, args ...interface{}) error {
    b := l.formatHeader(s, depth)
    fmt.Fprintln(b, args...)
    return l.write(s, b)
}

func (l *loggingT) Sync() {
    for range time.Tick(time.Second * 30) {
        l.sync()
    }
}

func (l *loggingT) sync() {
    l.mu.Lock()
    defer l.mu.Unlock()
    for i := 0; i < numSeverity; i++ {
        l.out[i].Sync()
    }
}

func (l *loggingT) info(depth int, args ...interface{}) error {
    return l.print(infoLog, depth, args...)
}

func (l *loggingT) Info(args ...interface{}) error {
    return l.info(1, args...)
}

func (l *loggingT) warning(depth int, args ...interface{}) error {
    return l.print(warningLog, depth, args...)
}

func (l *loggingT) Warning(args ...interface{}) error {
    return l.warning(1, args...)
}

func (l *loggingT) error(depth int, args ...interface{}) error {
    return l.print(errLog, depth, args...)
}
func (l *loggingT) Error(args ...interface{}) error {
    return l.error(1, args...)
}

func newLoggingT(name, dir string) (*loggingT, error) {
    res := &loggingT{}
    now := time.Now()
    for i := infoLog; i < numSeverity; i++ {
        file, err := createFile(name, dir, severityName[i], now)
        if err != nil {
            return nil, err
        }
        res.out[i] = file
    }
    go res.Sync()
    return res, nil
}

func NewLog(name, dir string) (Log, error) {
    return newLoggingT(name, dir)
}

var varLog *loggingT

var name = flag.String("log_name", "app", "the name of log")
var dir = flag.String("log_dir", os.TempDir(), "the directory of logs")

func Load() error {
    t, err := newLoggingT(*name, *dir)
    if err != nil {
        return err
    }
    varLog = t
    return nil
}

type Log interface {
    Info(...interface{}) error
    Warning(...interface{}) error
    Error(...interface{}) error
}

func Info(args ...interface{}) error {
    return varLog.info(1, args...)
}

func Warning(args ...interface{}) error {
    return varLog.warning(1, args...)
}

func Error(args ...interface{}) error {
    return varLog.error(1, args)
}
