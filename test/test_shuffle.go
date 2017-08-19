package main

import (
	"fmt"
	"math/rand"
	"time"
)

func randNum(low, high int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return low + r.Intn(high-low+1)
}

func shuffle(arr []int, n int) {

	if n <= 0 {
		return
	}

	shuffle(arr, n-1)
	rand := randNum(0, n)

	arr[n], arr[rand] = arr[rand], arr[n]

}

func main() {
	cards := []int{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
		14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
		25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35,
		36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46,
		47, 48, 49, 50, 51, 52}

	cardslen := len(cards)
	cardslen1 := cardslen / 4

	for i := 1; i <= 10; i++ {
		fmt.Printf("\n")
		shuffle(cards, cardslen-1)
		for j := 1; j <= cardslen; j++ {
			fmt.Printf("%d ", cards[j-1])
			if j%cardslen1 == 0 {
				fmt.Printf("\n")
			}
		}
	}
}

func Shuffle(vals []int) []int {  
  r := rand.New(rand.NewSource(time.Now().Unix()))
  ret := make([]int, len(vals))
  n := len(vals)
  for i := 0; i < n; i++ {
    randIndex := r.Intn(len(vals))
    ret[i] = vals[randIndex]
    vals = append(vals[:randIndex], vals[randIndex+1:]...)
  }
  return ret
}

func Shuffle(vals []int) []int {  
  r := rand.New(rand.NewSource(time.Now().Unix()))
  ret := make([]int, len(vals))
  perm := r.Perm(len(vals))
  for i, randIndex := range perm {
    ret[i] = vals[randIndex]
  }
  return ret
}

func main() {  
  vals := []int{10, 12, 14, 16, 18, 20}
  r := rand.New(rand.NewSource(time.Now().Unix()))
  for _, i := range r.Perm(len(vals)) {
    val := vals[i]
    fmt.Println(val)
  }
}

func Shuffle(vals []int) {  
  r := rand.New(rand.NewSource(time.Now().Unix()))
  for len(vals) > 0 {
    n := len(vals)
    randIndex := r.Intn(n)
    vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
    vals = vals[:n-1]
  }
}
