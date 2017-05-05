package main

import (
	"fmt"
	"strings"
	"strconv"
)

func isIpInRange(ip1 []int, ip2[]int) bool {
	for i:=0; i < 4; i++{
		if i < 3 {
			if ip1[i] < ip2[i] {
				fmt.Println("aaa: ", ip1, ip2)
				return false
			}
		} else {
			if ip1[i] <= ip2[i] {
				fmt.Println("bbb: ", ip1, ip2)	
				return false
			}
		}
	}
	for i:=0; i < 4; i++{
		if i < 3 {
			if ip1[i] > ip2[4+i] {
				fmt.Println("ccc: ", ip1, ip2)
				return false
			}
		} else {
			if ip1[i] >= ip2[4+i] {
				fmt.Println("ddd: ", ip1, ip2)	
				return false
			}
		}
	}
	return true
}

func IsInnerIP(ip string) bool {
	ipa := []int{10, 0, 0, 0, 10, 255, 255, 255}
	ipb := []int{172, 16, 0, 0, 172, 31, 255, 255}
	ipc := []int{192, 168, 0, 0, 192, 168, 255, 255}

	arr := strings.Split(ip, ".")
	if len(arr) == 4 {
		a0, err0 := strconv.Atoi(arr[0])
		a1, err1 := strconv.Atoi(arr[0])
		a2, err2 := strconv.Atoi(arr[0])
		a3, err3 := strconv.Atoi(arr[0])
		if err0 != nil && err1 != nil && err2 != nil && err3 != nil {
			fmt.Println(err0, err1, err2, err3)
			return false
		}
		ipi = []int{a0, a1, a2, a3}
		return isIpInRange(ipi, ipa) || isIpInRange(ipi, ipa) || isIpInRange(ipi, ipc)
	}
	
	return false
}

func main() {
	fmt.Println("192.168.1.188: ", IsInnerIP("192.168.1.188"))
	fmt.Println("10.66.69.48: ", IsInnerIP("10.66.69.48"))
	fmt.Println("172.18.69.78: ", IsInnerIP("172.18.69.78"))	
	fmt.Println("11.11.11.11: ", IsInnerIP("11.11.11.11"))	
}
