package log

import (
	"context"
	"testing"
	"time"
)

func TestInitDefault(t *testing.T) {
	options := &Options{
		Dir:    "./log/",
		Name:   "app",
		Level:  DEBUG,
		StdOut: true,
	}
	if err := InitDefault(options); err != nil {
		t.Fatal("create msg failed")
	}
	ctx := WithLogId(context.Background())
	for range 10 {
		ErrorF(ctx, "this is test, id:%d, name:%s", 12, "人呆君")
		Error(ctx, "some test", StringField("name", "kris"), IntField("age", 12), BoolField("flag", true))
		time.Sleep(time.Second)
	}
}
