package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Test struct {
	K uint
	V int
}
type Find struct {
	I int
	J int
	X int
}

var t5 = []Test{}
var t4 = []Test{}
var t3 = []Test{}
var t2 = []Test{}

var test = map[int][]Test{}

func main() {
	var x uint
	for x = 0; x < 32; x++ {
		t5 = append(t5, Test{x, 31 << x})
	}
	test[5] = t5
	for x = 0; x < 32; x++ {
		t4 = append(t4, Test{x, 15 << x})
	}
	test[4] = t4
	for x = 0; x < 32; x++ {
		t3 = append(t3, Test{x, 7 << x})
	}
	test[3] = t3
	for x = 0; x < 32; x++ {
		t2 = append(t2, Test{x, 3 << x})
	}
	test[2] = t2
	find := make([]Find, 10240)
	m := bits //getMetrix(20, 20)
	//printMetrix(m, nil, 0)
	start := time.Now()
	c := findAll(20, 20, m, find)
	end := time.Now()
	fmt.Println(c)
	//printMetrix(m, find, c)
	fmt.Println(len(m), end.Sub(start).Nanoseconds())
}

func findAll(i, j int, m []int, find []Find) int {
	findC := 0
	for x := 5; x > 1; x-- {
		for k := 0; k <= i-x; k++ {
			xc := 0xFFFFFFFF
			for l := 0; l < x; l++ {
				xc &= m[k+l]
			}
			for _, t := range test[x] {
				if xc&t.V == t.V {
					find[findC].I = k
					find[findC].J = int(t.K)
					find[findC].X = x
					for l := 0; l < x; l++ {
						m[k+l] &= ^t.V
					}
					findC++
					xc = 0xFFFFFFFF
					for l := 0; l < x; l++ {
						xc &= m[k+l]
					}
				}
			}
		}
	}

	return findC
}

var (
	bits = []int{
		1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
)

func getMetrix(i, j int) []int {
	m := make([]int, i)
	for k := 0; k < i; k++ {
		m[k] = 0x000FFFFF
		for n := 0; n < j; n++ {
			if rand.Int()%11 == 0 {
				m[k] &= ^(1 << uint(n))
			}
		}
		//for n := j; n < 32; n++ {
		//	m[k] &= ^(1 << uint(n))
		//}
	}
	return m
}

func printMetrix(m []int, find []Find, l int) {
	for k, n := range m {
		for i := 0; i < 20; i++ {
			s := n & (1 << uint(i))
			if s > 0 {
				s = 1
			}
			p := true
			if find != nil {
				for j := 0; j < l; j++ {
					if (k >= find[j].I) && (k < find[j].I+find[j].X) && (i >= find[j].J) && (i < find[j].J+find[j].X) {
						fmt.Printf("%8v", find[j].X)
						p = false
						break
					}
				}
			}
			if p {
				fmt.Printf("%8v", s)
			}
		}
		fmt.Println()
	}
	fmt.Println()
	fmt.Println()
}
