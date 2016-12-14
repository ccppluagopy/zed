package redis

import (
	"github.com/ccppluagopy/zed"
	redis "gopkg.in/redis.v5"
	"testing"
	"time"
)

var redisclusterpool *RedisClusterMgrPool

func SAdd(key string, members ...interface{}) bool {
	return redisclusterpool.DBAction(0, func(c *redis.ClusterClient) bool {

		ret := c.SAdd(key, members...)
		if ret.Err() != nil {
			zed.ZLog("redis sadd Error: %v", ret.Err())

			return false
		}

		zed.ZLog("redis sadd success: %v", ret.String())

		return true
	})
}

func Append(key, value string) bool {
	return redisclusterpool.DBAction(0, func(c *redis.ClusterClient) bool {

		ret := c.Append(key, value)
		if ret.Err() != nil {
			zed.ZLog("rediscluster sadd Error: %v", ret.Err())

			return false
		}

		zed.ZLog("rediscluster Append success: %v", ret.String())

		return true
	})
}

func GetSet(key string, value interface{}) (ret string, b bool) {
	b = redisclusterpool.DBAction(0, func(c *redis.ClusterClient) bool {

		val, err := c.GetSet(key, value).Result()
		if err != nil {
			zed.ZLog("GetSet Error: %v", err)

			return false
		}

		ret = val

		return true
	})

	return
}

func TestRedisCluster(t *testing.T) {
	redisclusterpool = NewRedisClusterMgrPool("rediscluster", []string{":7000", ":7001", ":7002"}, "", 2, 0)

	retvale, b := GetSet("db", "mongodb")
	if b {
		zed.ZLog("set key: db\n\tvale: %v", retvale)
	}
	//Append("db", "redis")

	retvale, b = GetSet("msg", "happy new year!")
	if b {
		zed.ZLog("set key: msg\n\tvale: %v", retvale)
	}
	//Append("msg", "haha")

	retvale, b = GetSet("fruits", "apple  apple !")
	if b {
		zed.ZLog("set key: fruits\n\tvale: %v", retvale)
	}
	//Append("fruits", "apple ")
	time.Sleep(240 * time.Second)
	//time.Sleep(30 * time.Second)
	Append("db", "redis")
	Append("msg", "haha")
	Append("fruits", "apple ")
	time.Sleep(10 * time.Second)

}
