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

func TestRedis() {
	redispool = zredis.NewRedisMgrPool("xx", "127.0.0.1:6379", 0, "", 1)
	path := "./test_redis.lua"
	mgr := redispool.GetMgr(0)
	sha1, _ := mgr.LoadScript(path)
	/*redispool.DBAction(0, func(c *redis.Client) bool {
		status := c.ScriptFlush()
		if status.Err() != nil {
			fmt.Println("Script Flush Error: ", status.Err())
			return false
		}
		fstr := readfile("./test_redis.lua")
		cmd := c.ScriptLoad(fstr)
		str, err := cmd.Result()
		fmt.Println("cmd 111: ", err, str)
		if err == nil {
			sha1 = str
		}
		return true
	})*/

	for i := 0; i < 10000; i++ {
		redispool.DBAction(0, func(c *redis.Client) bool {
			k, v := fmt.Sprintf("key_%d", i), fmt.Sprintf("data_%d", i)
			ret := c.EvalSha(sha1, []string{k}, v)
			str, err := ret.Result()
			retstr, ok := str.(string)
			fmt.Println("ret err: ", err, "ok:", ok, "str:", retstr)
			if err != nil {
				panic(err)
			}

			return true
		})
		time.Sleep(1 * time.Second)
	}

}

func main() {
	TestRedis()
}
