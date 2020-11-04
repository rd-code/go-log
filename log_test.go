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
	}
	if err := InitDefault(options); err != nil {
		t.Fatal("create msg failed")
	}
	Error("======", "1", 2, 3)
	Error("----------", 4, 5, 6)
	time.Sleep(time.Second)
}
