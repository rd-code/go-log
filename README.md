# 说明
golang 实现的log

# 使用示例

初始化

```go
	options := &log.Options{
		Dir:    "./log/",
		Name:   "app",
		Level:  DEBUG,
		StdOut: true,
	}
	log.InitDefault(options)
	log.Info(log.WithLogId(context.Background()),
             "================================",
             log.StringField("name", "test"))
```