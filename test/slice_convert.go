package main

import (
	"fmt"
	//"time"
	//"encoding/binary"
	"reflect"
	"runtime"
	"unsafe"
)

var (
	ch chan bool = make(chan bool, 10)
)

func xx() {
	for i := 0; i < 10; i++ {
		ch <- true
		fmt.Println("++ ", i)
	}
	close(ch)
}

func Slice(slice interface{}, newSliceType reflect.Type) interface{} {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Slice called with non-slice value of type %T", slice))
	}
	if newSliceType.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Slice called with non-slice type of type %T", newSliceType))
	}
	newSlice := reflect.New(newSliceType)
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(newSlice.Pointer()))
	hdr.Cap = sv.Cap() * int(sv.Type().Elem().Size()) / int(newSliceType.Elem().Size())
	hdr.Len = sv.Len() * int(sv.Type().Elem().Size()) / int(newSliceType.Elem().Size())
	hdr.Data = uintptr(sv.Pointer())
	return newSlice.Elem().Interface()
}

var buf1 []int32 = []int32{4}
var buf2 []byte = Slice(buf1, reflect.TypeOf([]byte)).([]byte)

func main() {

	var buf = make([]byte, 4)
	var f1 float32 = 3.14
	*((*float32)(unsafe.Pointer(&buf[0]))) = f1
	var f2 float32 = *((*float32)(unsafe.Pointer(&buf[0])))
	fmt.Println(buf)
	fmt.Println(f1, f2)

	buf = append(buf, "heheheheh"...)
	fmt.Println("0000", string(buf[4:]))

	for i := 0; i < 100; i++ {
		fmt.Println(i)
	}

	flag := testDefer()
	fmt.Println("testDefer: ", flag)

	var buf1 []int32 = []int32{4}
	var buf2 []byte = Slice(buf1, reflect.TypeOf([]byte)).([]byte)
	buf1 = append(buf1, buf3...)
	fmt.Println(buf1)
	fmt.Println(buf2)

}

type XXX struct {
	flag bool
	strs []string
}

func testDefer() *XXX {
	xxx := XXX{flag: false}

	defer func() {
		xxx.flag = true
		const maxStack = 20
		if err := recover(); err != nil {
			errstr := fmt.Sprintf("---------------------------------------\nruntime error: %v\ntraceback:\n", err)

			i := 1
			for {
				pc, file, line, ok := runtime.Caller(i)

				errstr += fmt.Sprintf("    stack: %d %v [file: %s] [func: %s] [line: %d]\n", i, ok, file, runtime.FuncForPC(pc).Name(), line)

				i++
				if !ok || i > maxStack {
					break
				}
			}
			errstr += "---------------------------------------\n"
			fmt.Println(errstr)
		}
	}()

	//panic("xxx")

	return &xxx
}
