/*
	go/src/time/sleep.go增加如下接口
	需要删除go/pkg对应的time库文件，再go install time才会生效
*/

func Async(f func()) {
	var t *Timer
	t = &Timer{
		r: runtimeTimer{
			when: when(1),
			f: func(arg interface{}, seq uintptr) {
				arg.(func())()
				t.Stop()
			},
			arg: f,
		},
	}
	startTimer(&t.r)
}

func AfterFuncWithoutGo(d Duration, f func()) *Timer {
	t := &Timer{
		r: runtimeTimer{
			when: when(d),
			f: func(arg interface{}, seq uintptr) {
				arg.(func())()
			},
			arg: f,
		},
	}
	startTimer(&t.r)
	return t
}
