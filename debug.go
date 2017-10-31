package zed

import (
	"fmt"
	//"os"
	"runtime"
	//"time"
)

const (
	maxStack  = 20
	separator = "---------------------------------------\n"
)

func HandlePanic(needLog bool, args ...interface{}) interface{} {
	if err := recover(); err != nil {
		errstr := fmt.Sprintf("%sruntime error: %v\ntraceback:\n", separator, err)

		i := 2
		for {
			pc, file, line, ok := runtime.Caller(i)

			if !ok || i > maxStack {
				break
			}

			errstr += fmt.Sprintf("    stack: %d %v [file: %s] [func: %s] [line: %d]\n", i-1, ok, file, runtime.FuncForPC(pc).Name(), line)

			i++
		}
		errstr += separator

		if needLog {
			ZLog(errstr)
		}

		if len(args) > 0 {
			cb, ok := args[0].(func())
			if ok {
				defer func() {
					recover()
				}()
				cb()
			}
		}
		return err
	}
	return nil
}

func GetStackInfo() string {
	errstr := fmt.Sprintf("%straceback:\n", separator)

	i := 2
	for {
		pc, file, line, ok := runtime.Caller(i)

		if !ok || i > maxStack {
			break
		}

		errstr += fmt.Sprintf("    stack: %d %v [file: %s] [func: %s] [line: %d]\n", i-1, ok, file, runtime.FuncForPC(pc).Name(), line)

		i++
	}
	errstr += separator

	return errstr
}

func LogStackInfo() {
	ZLog(GetStackInfo())
}
