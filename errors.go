package log

import "fmt"

var (
	UnknownLevelError = fmt.Errorf("unknown level error")
	DirectoryIsFile   = fmt.Errorf("the directory is file")
)
