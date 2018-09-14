package main

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
)

func isIpInRange(ip1 []int, ip2 []int) bool {
	for i := 0; i < 4; i++ {
		if i < 3 {
			if ip1[i] < ip2[i] {
				return false
			}
		} else {
			if ip1[i] <= ip2[i] {
				return false
			}
		}
	}
	for i := 0; i < 4; i++ {
		if i < 3 {
			if ip1[i] > ip2[4+i] {
				return false
			}
		} else {
			if ip1[i] >= ip2[4+i] {
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
		a1, err1 := strconv.Atoi(arr[1])
		a2, err2 := strconv.Atoi(arr[2])
		a3, err3 := strconv.Atoi(arr[3])
		if err0 != nil || err1 != nil || err2 != nil || err3 != nil {
			return false
		}
		return isIpInRange([]int{a0, a1, a2, a3}, ipa) || isIpInRange([]int{a0, a1, a2, a3}, ipb) || isIpInRange([]int{a0, a1, a2, a3}, ipc)
	}
	return false
}

func main() {
	ips := []string{
		"112.97.227.49",
		"124.167.36.15",
		"219.155.138.176",
		"223.104.90.11",
		"101.206.170.21",
		"61.158.149.251",
		"42.49.125.192",
		"113.121.42.15",
		"112.97.56.214",
		"222.184.112.145",
		"183.130.56.5",
	}
	for _, v := range ips {
		fmt.Println(v, ":", IsInnerIP(v))
	}

	// fmt.Println("192.168.1.188: ", IsInnerIP("192.168.1.188"))
	// fmt.Println("10.66.69.48: ", IsInnerIP("10.66.69.48"))
	// fmt.Println("172.18.69.78: ", IsInnerIP("172.18.69.78"))
	// fmt.Println("11.11.11.11: ", IsInnerIP("11.11.11.11"))
}
