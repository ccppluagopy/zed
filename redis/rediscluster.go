package redis

import (
	"sync"
	"time"

	"github.com/ccppluagopy/zed"

	redis "gopkg.in/redis.v5"
)

//---------------------------------------------------------cluster

type RedisClusterMgr struct {
	sync.RWMutex
	client   *redis.ClusterClient
	tryCount int
	addrs    []string
	passwd   string
	poolSize int
	ticker   *time.Ticker

	running    bool
	restarting bool
}

type RedisClusterMgrPool struct {
	redisClusterMgrs []*RedisClusterMgr
}

var (
	redisClusterMgrs     = make(map[string]*RedisClusterMgr)
	redisClusterMgrPools = make(map[string]*RedisClusterMgrPool)
)

// type ClusterOptions struct {
// 	Addrs []string
// 	MaxRedirects int
// 	Password string
// 	PoolSize           int
// }

func NewRedisClusterMgrPool(name string, addrs []string, passwd string, size int, connsize int) *RedisClusterMgrPool {
	if 0 == len(addrs) {
		zed.ZLog("NewRedisClusterMgrPool addrs is nil.")

		return nil
	}

	redisClusterMgrPool, isexist := redisClusterMgrPools[name]
	if isexist {
		zed.ZLog("NewRedisClusterMgrPool name(%s) redisClusterMgrPool has been existed.", name)

		return redisClusterMgrPool
	} else {
		redisClusterMgrPool = &RedisClusterMgrPool{
			redisClusterMgrs: make([]*RedisClusterMgr, size),
		}
	}

	if connsize <= 0 {
		connsize = 1
	}

	for index := 0; index < size; index++ {
		redisclustercopy := &RedisClusterMgr{
			tryCount:   0,
			addrs:      addrs,
			passwd:     passwd,
			poolSize:   connsize,
			running:    false,
			restarting: false,
		}

		ok := redisclustercopy.Start()
		if !ok {
			zed.ZLog("%s redisclustercopy.Start() %d Error.!", name, index)

			return nil
		}

		redisClusterMgrPool.redisClusterMgrs[index] = redisclustercopy
	}

	return redisClusterMgrPool
}

func (redisclustermgr *RedisClusterMgr) Start() bool {
	if !redisclustermgr.IsRuning() {
		clusterclient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: redisclustermgr.addrs,
			//MaxRedirects int
			Password: redisclustermgr.passwd,
			PoolSize: redisclustermgr.poolSize,
		})

		_, err := clusterclient.Ping().Result()
		if err != nil {
			zed.ZLog("RedisClusterMgr Start err: %v addr: %s", err, redisclustermgr.addrs)
			if redisclustermgr.tryCount < DB_DIAL_MAX_TIMES {
				redisclustermgr.tryCount = redisclustermgr.tryCount + 1
				time.Sleep(time.Second * 4)

				return redisclustermgr.Start()
			} else {

				return false
			}
		}

		redisclustermgr.client = clusterclient
		redisclustermgr.tryCount = 0
		redisclustermgr.ticker = time.NewTicker(10 * time.Second)
		redisclustermgr.SetRunningState(true)
		redisclustermgr.restarting = false

		zed.NewCoroutine(func() {
			redisclustermgr.startHeartBeat()
		})

		zed.ZLog("RedisClusterMgr add； %s Start---->>>", redisclustermgr.addrs)
	}

	return true
}

//GetMgr ...
func (pool *RedisClusterMgrPool) GetClusterMgr(idx int) *RedisClusterMgr {

	return pool.redisClusterMgrs[idx%len(pool.redisClusterMgrs)]
}

//DBAction ...
func (pool *RedisClusterMgrPool) DBAction(idx int, cb func(*redis.ClusterClient) bool) bool {

	return pool.GetClusterMgr(idx).DBAction(cb)
}

//Client ...
func (pool *RedisClusterMgrPool) ClusterClient(idx int) *redis.ClusterClient {

	return pool.GetClusterMgr(idx).client
}

//NewRedisClusterMgr ...
func NewRedisClusterMgr(name string, addrs []string, passwd string, connsize int) *RedisClusterMgr {
	redisclustermgr, isexist := redisClusterMgrs[name]
	if !isexist {
		redisclustermgr = &RedisClusterMgr{
			client:     nil,
			tryCount:   0,
			addrs:      addrs,
			passwd:     passwd,
			poolSize:   connsize,
			running:    false,
			restarting: false,
		}

		ok := redisclustermgr.Start()
		if !ok {
			zed.ZLog("NewRedisClusterMgr redismgr.Start() Error")

			return nil
		}

		return redisclustermgr
	}
	zed.ZLog("NewRedisClusterMgr Error: %s has been existed!", name)

	return nil
}

//DBAction ...
func (redisclustermgr *RedisClusterMgr) DBAction(cb func(*redis.ClusterClient) bool) bool {
	defer func() {
		if err := recover(); err != nil {
			zed.LogStackInfo()
			redisclustermgr.Restart()
		}
	}()

	client := redisclustermgr.client
	if client != nil {

		return cb(client)
	} else {

		return false
	}

	return true
}

//Restart ...
func (redisclustermgr *RedisClusterMgr) Restart() {
	zed.ZLog("enter Restart function ...")
	redisclustermgr.Lock()

	if !redisclustermgr.restarting {
		redisclustermgr.restarting = true
		redisclustermgr.Unlock()
		redisclustermgr.Stop()
		redisclustermgr.Start()
	}
}

//Stop ...
func (redisclustermgr *RedisClusterMgr) Stop() {
	zed.ZLog("enter Stop function ...")
	redisclustermgr.SetRunningState(false)

	redisclustermgr.Lock()
	defer redisclustermgr.Unlock()
	//zed.ZLog("return back Stop function ...")

	err := redisclustermgr.client.Close()
	if err != nil {
		zed.ZLog("RedisClusterMgr stop client() err: %v", err)
		panic(err)
	}
}

//startHeartBeat ...
func (redisclustermgr *RedisClusterMgr) startHeartBeat() {
	for {
		select {
		case _, ok := <-redisclustermgr.ticker.C:
			if ok {
				redisclustermgr.heartBeat()
			} else {

				break
			}
		}
	}
}

//heartBeat ...
func (redisclustermgr *RedisClusterMgr) heartBeat() {
	zed.ZLog("enter heartbeat function....")
	redisclustermgr.Lock()
	//redismgr.Unlock()

	if redisclustermgr.client != nil {
		zed.ZLog("redisclustermgr.client is not nil.")
		_, err := redisclustermgr.client.Ping().Result()
		if err != nil {
			zed.ZLog("RedisClusterMgr heartBeat err: %v!", err)
			//panic(err)
			redisclustermgr.Unlock()
			redisclustermgr.Restart()
		} else {
			zed.ZLog("redisclustermgr.client online.")
			redisclustermgr.Unlock()
		}
	}
}

//IsRuning ...
func (redisclustermgr *RedisClusterMgr) IsRuning() bool {
	redisclustermgr.Lock()
	defer redisclustermgr.Unlock()

	return redisclustermgr.running
}

func (redisclustermgr *RedisClusterMgr) SetRunningState(b bool) {
	//zed.ZLog("enter SetRunningState function ...")
	redisclustermgr.Lock()
	defer redisclustermgr.Unlock()

	redisclustermgr.running = b
	//ed.ZLog("change running state ...")
}

//GetRedisMgrByName ...
func GetRedisClusterMgrByName(name string) *RedisClusterMgr {
	redisclustermgr, isexist := redisClusterMgrs[name]
	if !isexist {
		zed.ZLog("GetRedisClusterMgrByName Error: %s is not exist!", name)

		return nil
	}

	return redisclustermgr
}

//GetRedisMgrPoolByName ...
func GetRedisCLusterMgrPoolByName(name string) (*RedisClusterMgrPool, bool) {
	mgr, isexist := redisClusterMgrPools[name]

	return mgr, isexist
}
