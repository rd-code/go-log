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
		StdOut:   false,
		Current:  20,
	}
	if err := InitDefault(options); err != nil {
		t.Fatal("create msg failed")
	}
	for {
		Error("this is test, id:%d, name:%s", 12, "人呆君")
		time.Sleep(time.Second)
	}
}
