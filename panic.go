package zed

import (
	"fmt"
	"runtime"
)

const (
	maxStack  = 20
	separator = "---------------------------------------\n"
)

func PanicHandle() {
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

		//LogError(LogConf.Panic, LogConf.SERVER, errstr)
	}
}
