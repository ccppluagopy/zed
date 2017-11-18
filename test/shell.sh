#!/usr/bin/env bash

name="admin"
pass=$(echo -n 'admin' | md5sum | cut -d ' ' -f 1)
group="superuser"

mongo $* <<EOF

进程内打开的文件描述符数量
ps -ef | grep xxx.bin
lsof -p xxx.bin_pid | wc -l
