指标参考下面几个内容整理。
- basic-perfomance
- tsar https://github.com/alibaba/tsar/blob/master/info.md#io
- netflixs http://techblog.netflix.com/2015/04/introducing-vector-netflixs-on-host.html
- openfalcon http://book.open-falcon.com/zh/faq/linux-metrics.html

含义参考了: 对应的 manpage,  [kernel doc](https://www.kernel.org/doc/Documentation/filesystems/proc.txt)。

## 系统监控

### 目标

- 每一次系统层面发生的故障，监控的报警都要能发现。

- 系统采集的指标，应该能表明系统在某个时刻是不是有问题。

### 监控项
#### uptime

``` 
cat /proc/uptime
5536402.56 5506312.54
```

- 第一列启动时间
- 第二列是系统空闲(idle)时间

单位是 `s`。

#### cpu

可以从 `/proc/stat`，也可以从 `top` 来看到这些指标的值。

``` 
cpu  2141528 0 556623 548389457 406368 115 10325 116432 0 0
cpu0 2141528 0 556623 548389457 406368 115 10325 116432 0 0
intr 551323037 19 9 0 0 2 0 2 0 2 0 0 0 144 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 474757771 0 0 0 0 510 0 4150701 72394413 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 ....(好多〇)
ctxt 1870344087
btime 1447738559
processes 2930163
procs_running 1
procs_blocked 0
softirq 224417040 0 77979395 36732867 41760790 0 0 133963 0 81548 67728477
```

- **cpu**: 单位是`jiffies`[^1]。挨个是：
  - user: normal processes executing in user mode。这个值应该越高越好。
  - nice: niced processes executing in user mode。通常我们不会有。
  - system: processes executing in kernel mode。一般的不希望这个值太高。
  - idle: twiddling thumbs
  - iowait: waiting for I/O to complete
  - irq: servicing interrupts
  - softirq: servicing softirqs
  - steal: involuntary wait. 被强制等待虚拟CPU的时间,此时hypervisor在为另一个虚拟处理器服务。这个值高了表示物理机超卖了。
  - guest: running a normal guest, 运行guest os 虚拟 cpu 所用。
  - guest_nice: running a niced guest
- intr 是总的中断的统计。
- ctxt 所有 cpu 上的 context switches 总数。
- btime system boot time, 单位 seconds, 从 unix epoch 开始算。
- processes 表示创建过的`processes and threads`， 包括但不限于 fork() clone()创建的。



所以在 cpu 需要关注的指标有 `cpu.user`, `cpu.sys`, `cpu.idel`, `cpu.iowait`, `cpu.irq`, `cpu.softirq`, `cpu.util`, `cpu.switches`.

可以这样算 cpu.util:  1 - idle - iowait - steal

[^1]: jiffies 的大小在  `<linux/jiffies.h> `里定义。x86平台上是 1/100s。

#### load

从 `/proc/loadavg` 看。

``` 
cat /proc/loadavg
0.00 0.02 0.05 1/186 17582
```

分别是 过去一分钟负载，过去5分钟负载，过去15分钟负载，运行进程/总进程数，最大的 pid。



load 可以采集 load.1min, load.5min, load.15min, load.runq(即上面的 process_running)。



#### 内存

内存的指标比较多，要判断一个机器内存使用的状态，要采集这几个：

- total： 总的物理内存。
- used：使用的大小。
- buff：buffer is something that has yet to be written to disk。
- cache
- util： (total - free - buff - cache) / total * 100%
- swap total
- swap used
- swap util: swap used / swap total * 100%.



平时关心 util 和 swap util 就能直观的判断内存的负载情况。



page fault 也可以采集一下。

- major pgfault，分配了虚拟内存地址，但是对应的物理内存地址没分配、或者内容不存在的时候，需要分配内存、从磁盘读数据。
- minor pgfault，因为各种原因（read-only）之类，进程不能对内存页做操作的时候发生。

如何查看：

``` 
grep pgfault /proc/vmstat
cut -d " " -f 1,2,10,12 /proc/pid/stat
```



#### I/O

数据都在 `/proc/diskstats`里面。我们平时都看 `iostat -x 1` 的输出。先看下它的输出的含义(翻译自 manpage)：

- rrqm/s: 每秒加到读队列的读操作数量。
- wrqm/s: 每秒加到写队列的读操作数量。
- r/s: 每秒完成读操作的次数。
- w/s: 每秒完成写操作的次数。 

**上面说的操作都是 merge 之后的，相邻的读/写可能会合并。**

**rrqm是系统合并后的值， r 是真正落到磁盘的上请求数量。**

- rsec/s: 每秒读的扇区数。
- wsec/s: 每秒写的扇区数。
- rkB/s:  每秒读K字节数.是 rsect/s 的一半,因为每扇区大小为512字节。
- wkB/s:  每秒读K字节数.是 rsect/s 的一半,因为每扇区大小为512字节。
- avgrq-sz：平均每次I/O操作的数据大小 (扇区)。
- avgqu-sz：平均I/O队列长度。.
- await: 平均每次操作的等待时间。 r_await 和 w_await 解释和这个一模一样。
- svctm: manpage 说这个指标要废弃了。
- util: 理解成采样时间内有多少时间队列是非空的。这个值太高说明磁盘存在瓶颈。



这些指标都可以采集回来。

平时我们重点关注 r/s w/s 和 util 就够了。 前两个反映读写的 iops，最后一个反映磁盘的整体负载情况。



本节参考：

- http://linux.die.net/man/1/iostat
- [https://github.com/alibaba/tsar/blob/master/info.md#io](https://github.com/alibaba/tsar/blob/master/info.md#io)


#### 分区

采集和关注这几个指标：

- used
- free
- inode used
- inode free
- total
- rw： 分区是否可读写。

要排除掉docker 产生的 union 分区。



#### 网卡

这些指标是网卡的统计数据，可以用 `ethtool -S ethx` 看到。数据源来自：`/proc/net/dev`和 `/sys/class/net/*/statistics/*`。



- net.if.in.bytes : 收到的数据的总 bytes.


- net.if.in.dropped: 进入了 ring buffer， 在拷贝到内存的时候被丢了。


- net.if.in.errors: 收到的错误包的总数。错误包包括：crc 校验错误、帧同步错误、ring buffer溢出等。


- net.if.in.fifo.errs: 这个是 ifconfig 里看到的 overruns.表示数据包没到 ring buffer 就被丢了。也就是 cpu 来不及处理 ringbuffer 里的数据，通常在网卡压力大、没有做affinity的时候会发生。


- net.if.in.frame.errs: misaligned frames. frame 的长度（bit）不能被 8 整除。


- net.if.in.multicast: 组播。


- net.if.in.packets: 收到的 packets 数量统计。


- net.if.out.bytes


- net.if.out.carrier.errs: 这个意味着物理层出问题了。比如网卡的工作模式不对。


- net.if.out.collisions： 因为 CSMA/CD 造成的传输错误。


- net.if.out.dropped


- net.if.out.errors


- net.if.out.fifo.errs


- net.if.out.packets


- net.if.total.bytes


- net.if.total.dropped


- 网卡的工作模式。全双工千兆/万兆。

网络压力大的服务，是有必要做中断的 affinity 和提高 ring buffer 值的。



#### 网络

从 `/proc/net/snmp` 可以拿到这些指标：

**tcp**

- ActiveOpens:主动打开的tcp连接数量。
- PassiveOpens:被动打开的tcp连接数量。
- InSegs: 收到的tcp报文数量。
- OutSegs:发出的tcp报文数量。
- EstabResets: established 中发生的 reset。
- AttemptFails: 连接失败的数量。
- CurrEstab:当前状态为ESTABLISHED的tcp连接数。
- RetransSegs: 重传的报文数量。
- retran:系统的重传率 (RetransSegs－last RetransSegs) ／ (OutSegs－last OutSegs) * 100%

**udp**

- InDatagrams
- OutDatagrams
- NoPorts: 目的地址或者端口不存在。
- InErrors： 无效数据包。
- RcvbufErrors：内核的 buffer 满了导致的接收失败。
- SndbufErrors：同上。
- InCsumErrors：checksum 错误的 udp 包数量。

还可以用 ss 把这些 tcp 链接的信息采集起来：

- ss.orphaned
- ss.closed
- ss.timewait
- ss.slabinfo.timewait
- ss.synrecv
- ss.estab

collectd 默认会把每个 TCP 状态的连接数都采集下来。

这些指标可能不需要报警。但是收集起来用来排查问题、查看网络负载很有用。



### sysctl

有些 sysctl 的配置项的值也需要关心下：

- net.netfilter.nf_conntrack_max
- net.netfilter.nf_conntrack_count
- fs.file-max 、fs.file-nr 这两个参数可能更需要关注单个进程内的值。

还有一些我们自己调整过的、对业务会有影响的项。比如 ip_forward.

### 硬件层面

- 电源健康状态。是不是两路电都在。

- 硬盘的健康状态。 分区是否可写。

- raid 卡的健康状态。

- 远控卡是不是可用。

- 远控卡中的异常日志。

- 网卡的工作状态是否正常。  






