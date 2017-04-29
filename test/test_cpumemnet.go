package main

import(
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"time"
)

func main() {
	fmt.Println("----------------------")
	v, _ := mem.VirtualMemory()
	fmt.Printf("Total: %v, Free: %v, Used: %v, UsedPercent: %v\n", v.Total, v.Free, v.Used, v.UsedPercent)
	
	fmt.Println("----------------------")

	percents, _ := cpu.Percent(time.Microsecond, true)
	infos, _ := cpu.Info()
	ncpu, _ := cpu.Counts(true)
	timestats, _ := cpu.Times(true)
	for i:=0; i < len(percents); i++ {
		fmt.Printf("cpu %d UsedPercent: %v\n", i, percents[i])
	}
	fmt.Printf("ncpu: %d\n", ncpu)
	fmt.Println("infos: ", infos)
	fmt.Println("timestats: ", timestats)
	
	fmt.Println("----------------------")
	fmt.Println("----------------------")
	fmt.Println("----------------------")
	
	msize = 1024*1024
	prenetstats, _ := net.IOCounters(true)
	for {
		time.Sleep(time.Second)
		netstats, _ := net.IOCounters(true)
		fmt.Printf("%s Recv Speed: %d M/s, Send Speed: %d M/s\n",
			netstats[i].Name,
			(netstats[i].ByteRecv-prenetstats[i].BytesRecv)/msize,
			(netstats[i].ByteSent-prenetstats[i].BytesSent)/msize)
		prenetstats = netstats
		fmt.Println("----------------------")
	}
}
