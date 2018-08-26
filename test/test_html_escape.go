package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"strings"
)

//HTMLEscape 函数将添加Buffer中的特殊字符串进行转义(Buffer中本来就有的字节不会转义只会转义后来添加的)
//Compact 对这些特殊字符不进行转义 但是有一个作用就是在拼接字符串时 如果后面的字符串有问题那么不添加到Buffer中相当于自动帮我们检查了
//Compact 是个很有用的函数试想一下如果不用Compact直接拼接字符串那么如果其中第n个json串格式有问题就会导致后面所有的json串无法解析
//而使用了Compact来解析 假如拼接到第n个json串有问题会直接抛弃这个json串而不会去影响到后面json串的解析
//比如例子中errJson字符串的格式是有问题的 通过Compact会自动检查所以最后输出的结果是不包含errJson的
//特殊字符都有<, >, &, U+2028 and U+2029 转义成  \u003c, \u003e, \u0026, \u2028, \u2029
func main() {
	buf := bytes.NewBufferString("")
	str := `{"Name":"<wujunbin>", "Age":21}`
	errJson := `{Name:"<wujunbin>", "Age":21}`
	//拼接json串
	fmt.Println("-- 111:", json.Compact(buf, []byte(str)))
	fmt.Println("-- 222:", json.Compact(buf, []byte(errJson)))
	bufEscape := bytes.NewBufferString("")
	//拼接转义的json串  但是不会帮你检查错误
	json.HTMLEscape(bufEscape, buf.Bytes())
	//result is {"Name":"\u003cwujunbin\u003e","Age":21}
	fmt.Printf("-- 333: %s \n", string(bufEscape.Bytes()))

	type Person struct {
		Name string
		Age  int
	}
	decoder := json.NewDecoder(strings.NewReader(string(buf.Bytes())))
	for {
		var p Person
		if err := decoder.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("-- 444:", err)
			break
		}
		fmt.Println("-- 555:", p)
	}
	fmt.Println("-- 666:", html.EscapeString(str))
	fmt.Println("-- 777:", html.EscapeString(errJson))
}
