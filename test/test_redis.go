package main

import (
	"fmt"
	zredis "github.com/ccppluagopy/zed/redis"
	redis "gopkg.in/redis.v5"
	//"testing"
	"time"
)

var redispool *zredis.RedisMgrPool

func TestRedisCluster() {
	redispool = zredis.NewRedisMgrPool("xx", "127.0.0.1:6379", 0, "", 3)
	redispool.DBAction(0, func(c *redis.Client) bool {
		ret := c.Set("key", "hahaha", 0)
		c.Set(key, value, expiration)
		ret2 := c.Get("key")
		fmt.Println("ret: ", ret)
		fmt.Println("ret2: ", ret2)
		return true
	})
	time.Sleep(10 * time.Second)

}

func main() {
	TestRedisCluster()
}
