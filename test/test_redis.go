package main

import (
	"fmt"
	zredis "github.com/ccppluagopy/zed/redis"
	redis "gopkg.in/redis.v5"
	//"testing"
	"io/ioutil"
	"os"
	"time"
)

var redispool *zredis.RedisMgrPool

func readfile(path string) string {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	buf, err := ioutil.ReadAll(fi)
	// fmt.Println(string(fd))
	return string(buf)
}

func TestRedisCluster() {
	redispool = zredis.NewRedisMgrPool("xx", "127.0.0.1:6379", 0, "", 3)
	var cmd *redis.StringCmd

	redispool.DBAction(0, func(c *redis.Client) bool {
		status := c.ScriptFlush()
		if status.Err() != nil {
			fmt.Println("Script Flush Error: ", status.Err())
			return false
		}
		fstr := readfile("./redis.lua")
		//fmt.Println("file: ", fstr)
		cmd = c.ScriptLoad(fstr)
		err, str := cmd.Result()
		fmt.Println("cmd 111: ", err, str)
		return true
	})

	for i := 0; i < 10000; i++ {
		redispool.DBAction(0, func(c *redis.Client) bool {

			ret := c.EvalSha(cmd.Val(), []string{"gaofeng", "gaofeng"}, "gaofeng")

			str, err := ret.Result()
			retstr, ok := str.(string)
			fmt.Println("ret err: ", err, "ok:", ok, "str:", retstr)
			//fmt.Println("ret2: ", ret2.String())
			return true
		})
		time.Sleep(10 * time.Second)
	}

}

func main() {
	TestRedisCluster()
}
