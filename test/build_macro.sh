#!/bin/bash


BUILD_TIME=`date +%FT%T%z`
SVN_VER=$(svn info | grep "Rev:" | cut -d ' ' -f 4)
VERSION=$(echo $BUILD_TIME$SVN_VER)
go build -ldflags "-X main.VERSION=${VERSION}" -o ./main.bin main.go



mkdir bin
cp config.json start.sh stop.sh bin/
zip bin.zip bin -r
rm -rf bin 

#windows
#set BUILD_TIME=%date:~0,4%-%date:~5,2%-%date:~8,2%T%time:~0,8%
#go build -ldflags "-X main.BuildTime=%BUILD_TIME%" -o ./main.bin main.go
