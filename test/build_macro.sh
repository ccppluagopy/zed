#!/bin/bash

BUILD_TIME=`date +%FT%T%z`
go build -ldflags "-X main.BuildTime=${BUILD_TIME}" -o ./main.bin main.go

mkdir bin
cp config.json start.sh stop.sh bin/
zip bin.zip bin -r
rm -rf bin 

#windows
#set BUILD_TIME=%date:~0,4%-%date:~5,2%-%date:~8,2%T%time:~0,8%
#go build -ldflags "-X main.BuildTime=%BUILD_TIME%" -o ./main.bin main.go
