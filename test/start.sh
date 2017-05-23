#!/bin/bash

ulimit -n 65535

sysctl -w net.ipv4.tcp_keepalive_intvl=75
sysctl -w net.ipv4.tcp_keepalive_probes=9
sysctl -w net.ipv4.tcp_keepalive_time=7200
sysctl -w net.ipv4.tcp_max_syn_backlog=65536
sysctl -w net.ipv4.tcp_max_tw_buckets=65536=75
sysctl -w net.ipv4.tcp_syn_retries=6
sysctl -w net.ipv4.tcp_synack_retries=5
sysctl -w net.ipv4.tcp_syncookies=1
sysctl -w net.ipv4.tcp_tw_recycle=1
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -p

DIR=`dirname $0`
cd $DIR
if [ ! -d "./logs" ]; then
  mkdir ./logs
fi

rm -rf ./nohup.out >/dev/null 2>/dev/null
PROG=./goproxy.bin
chmod a+x $PROG
echo "start"
nohup $PROG &
