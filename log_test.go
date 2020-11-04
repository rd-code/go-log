package log

import (
	"testing"
	"time"
)

func TestInitDefault(t *testing.T) {
	options := &Options{
		Dir:      "./log/",
		Name:     "app",
		Severity: DEBUG,
		StdOut:   true,
		Current:  20,
	}
	if err := InitDefault(options); err != nil {
		t.Fatal("create msg failed")
	}
	Error("======", "1", 2, 3)
	Error("----------", 4, 5, 6)
	Error("------", 5)
	ErrorF("id:%d, name:%s", 15, "20")
	time.Sleep(time.Second)
}
