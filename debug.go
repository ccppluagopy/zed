package zed

import (
	"fmt"
	"runtime"
)

const (
	maxStack  = 20
	separator = "---------------------------------------\n"
)

func PanicHandle(needLog bool, args ...interface{}) {
	if len(args) > 0 {
		LogInfo(LOG_IDX, LOG_IDX, args[0])
	}

	if err := recover(); err != nil {
		errstr := fmt.Sprintf("%sruntime error: %v\ntraceback:\n", separator, err)

		i := 1
		for {
			pc, file, line, ok := runtime.Caller(i)

			errstr += fmt.Sprintf("    stack: %d %v [file: %s] [func: %s] [line: %d]\n", i, ok, file, runtime.FuncForPC(pc).Name(), line)

			i++
			if !ok || i > maxStack {
				break
			}
		}
		errstr += separator

		if needLog {
			LogError(LOG_IDX, LOG_IDX, errstr)
		}
	}
}

func GetStackInfo() string {
	errstr := fmt.Sprintf("%straceback:\n", separator)

	i := 1
	for {
		pc, file, line, ok := runtime.Caller(i)

		errstr += fmt.Sprintf("    stack: %d %v [file: %s] [func: %s] [line: %d]\n", i, ok, file, runtime.FuncForPC(pc).Name(), line)

		i++
		if !ok || i > maxStack {
			break
		}
	}
	errstr += separator

	return errstr
}
