#!/bin/bash

BUILD_TIME=`date +%FT%T%z`
go build -ldflags "-X main.BuildTime=${BUILD_TIME}" -o ./main.bin main.go

mkdir bin
cp config.json start.sh stop.sh bin/
zip bin.zip bin -r
rm -rf bin 
