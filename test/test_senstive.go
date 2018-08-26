package main

import (
	"bufio"
	"fmt"
	"github.com/armon/go-radix"
	"log"
	"os"
	"strings"
)

var (
	//_dirtyCharacters = "`~!@#$%^&*()_+{}|\\'\";:,./<>?-= \t\n，。？、‘；：【】{}~！@#￥%……&*（）·——+=-、|"
	_dirtyCharacters = "<>,;\\"
	_dirtyWords      = []string{
		"script",
		"select",
	}
)
var (
	_radixTreeRoot *radix.Tree = radix.New()
)

func initSensetiveWord() {
	var (
		err  error
		file *os.File
		part []byte
	)

	if file, err = os.Open("./conf/sensetive.txt"); err != nil {
		panic(err)
	}

	fmt.Println("file, err:", file, err)

	reader := bufio.NewReader(file)

	n := 0
	for _, c := range _dirtyCharacters {
		n++
		_radixTreeRoot.Insert(string(c), n)
	}
	for _, str := range _dirtyWords {
		n++
		_radixTreeRoot.Insert(str, n)
	}
	for {
		part, _, err = reader.ReadLine()
		if err != nil {
			break
		}

		word := strings.Trim(strings.Trim(strings.Trim(string(part), " "), "	"), "\n")
		if len(word) > 0 {
			n++
			_radixTreeRoot.Insert(strings.Trim(strings.Trim(strings.Trim(string(part), " "), "	"), "\n"), n)
		}

	}
}

func isValidString(str string, args ...interface{}) bool {
	for i := 0; i < len(str); i++ {
		tmp, inter, ok := _radixTreeRoot.LongestPrefix(str[i:])
		if len(args) > 0 {
			fmt.Printf("xxx: '%v', %v, %v, %v\n", tmp, inter, ok, str[i:])
		}
		if ok {
			return false
		}
	}
	return true
}

func test2() {
	str := "aaa毛李大钊"
	if isValidString(str) {
		log.Fatalf("TestSensetiveWord failed: %s", str)
	}

	str = "陈云bbb"
	if isValidString(str) {
		log.Fatalf("TestSensetiveWord failed: %s", str)
	}

	str = "aaa专政bbb"
	if isValidString(str) {
		log.Fatalf("TestSensetiveWord failed: %s", str)
	}

	str = "aa暴力虐待ddd"
	if isValidString(str) {
		log.Fatalf("TestSensetiveWord failed: %s", str)
	}

	str = "是是是毛天下泽哈哈东"
	if isValidString(str, true) {
		log.Fatalf("TestSensetiveWord failed: %s", str)
	}

	str = "aaa毛李大cccc 钊"
	if isValidString(str) {
		log.Fatalf("TestSensetiveWord failed: %s", str)
	}

	str = "陈a云bbb"
	if isValidString(str) {
		log.Fatalf("TestSensetiveWord failed: %s", str)
	}

	str = "aaa专nnn政bbb"
	if isValidString(str) {
		log.Fatalf("TestSensetiveWord failed: %s", str)
	}

	str = "aa暴h力虐nnnnn待ddd"
	if isValidString(str) {
		log.Fatalf("TestSensetiveWord failed: %s", str)
	}
}

func main() {
	initSensetiveWord()

	str := "小鱼select"
	// if !isValidString(str) {
	// 	log.Fatalf("TestSensetiveWord failed: %s", str)
	// }
	fmt.Println("result:", isValidString(str))
}
