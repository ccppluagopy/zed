package redis

import (
	"sync"
	"time"

	"github.com/ccppluagopy/zed"
	redis "gopkg.in/redis.v5"
	"io/ioutil"
	"os"
)

const (
	DB_DIAL_TIMEOUT   = time.Second * 10
	DB_DIAL_MAX_TIMES = 1000
)

type RedisActionCB func(redis *RedisMgr) bool

type RedisMgr struct {
	sync.RWMutex
	client      *redis.Client
	tryCount    int
	addr        string
	database    int
	passwd      string
	poolSize    int
	ticker      *time.Ticker
	running     bool
	restarting  bool
	scriptCache map[string]string
}

type RedisMgrPool struct {
	redismgrs []*RedisMgr
}

var (
	redisMgrs     = make(map[string]*RedisMgr)
	redisMgrPools = make(map[string]*RedisMgrPool)
)

//NewRedisMgrPool ...
func NewRedisMgrPool(name, addr string, database int, passwd string, size int) *RedisMgrPool {
	redismgrpool, isexist := redisMgrPools[name]
	if !isexist {
		redismgrpool = &RedisMgrPool{
			redismgrs: make([]*RedisMgr, size),
		}
	}

	var ok bool
	for index := 0; index < size; index++ {
		rediscopy := &RedisMgr{
			tryCount:    0,
			addr:        addr,
			database:    database,
			passwd:      passwd,
			poolSize:    1,
			running:     false,
			restarting:  false,
			scriptCache: make(map[string]string),
		}

		ok = rediscopy.Start()
		if !ok {
			zed.ZLog("%s rediscopy.Start() %d Error.!", name, index)

			return nil
		}

		redismgrpool.redismgrs[index] = rediscopy
	}

	return redismgrpool
}

//GetMgr ...
func (pool *RedisMgrPool) GetMgr(idx int) *RedisMgr {

	return pool.redismgrs[idx%len(pool.redismgrs)]
}

func (pool *RedisMgrPool) GetScriptSha1(path string) string {
	return pool.redismgrs[0].GetScriptSha1(path)
}

//DBAction ...
func (pool *RedisMgrPool) DBAction(idx int, cb func(*redis.Client) bool) bool {

	return pool.GetMgr(idx).DBAction(cb)
}

//Client ...
func (pool *RedisMgrPool) Client(idx int) *redis.Client {

	return pool.GetMgr(idx).client
}

//NewRedisMgr ...
func NewRedisMgr(name, addr string, database int, passwd string) *RedisMgr {
	// database := append(database, 0)
	// passwd := append(passwd, "")

	redismgr, isexist := redisMgrs[name]
	if !isexist {
		redismgr = &RedisMgr{
			client:      nil,
			tryCount:    0,
			addr:        addr,
			database:    database,
			passwd:      passwd,
			running:     false,
			restarting:  false,
			scriptCache: make(map[string]string),
		}

		ok := redismgr.Start()
		if !ok {
			zed.ZLog("NewRedisMgr redismgr.Start() Error")

			return nil
		}

		return redismgr
	}
	zed.ZLog("NewRedisMgr Error: %s has been existed!", name)

	return nil
}

func (redismgr *RedisMgr) GetScriptSha1(path string) string {
	if sha1, ok := redismgr.scriptCache[path]; ok {
		return sha1
	}
	return ""
}

func (redismgr *RedisMgr) LoadScript(path string) (string, bool) {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	buf, err := ioutil.ReadAll(fi)
	// fmt.Println(string(fd))
	fstr := string(buf)
	ret := redismgr.client.ScriptLoad(fstr)
	str, err := ret.Result()
	if err != nil {
		zed.ZLog("RedisMgr LoadScript Error: %s %s %v", path, str, err)
		return "", false
	}
	zed.ZLog("RedisMgr LoadScript %s", path)
	redismgr.scriptCache[path] = str
	return str, true
}

//Start ...
func (redismgr *RedisMgr) Start() bool {
	if !redismgr.IsRuning() {
		zed.ZLog("RedisMgr Start - 00000")
		client := redis.NewClient(&redis.Options{
			Addr:        redismgr.addr,
			Password:    redismgr.passwd,
			DB:          redismgr.database, /*连接的表的编号*/
			PoolSize:    redismgr.poolSize, /**/
			DialTimeout: time.Second / 20,
		})
		_, err := client.Ping().Result()
		if err != nil {
			zed.ZLog("RedisMgr Start err: %v addr: %s database: %d", err, redismgr.addr, redismgr.database)
			if redismgr.tryCount < DB_DIAL_MAX_TIMES {
				redismgr.tryCount = redismgr.tryCount + 1
				time.Sleep(time.Second * 1)

				return redismgr.Start()
			} else {

				return false
			}
		}

		redismgr.client = client
		redismgr.tryCount = 0
		redismgr.ticker = time.NewTicker(10 * time.Second)
		redismgr.SetRunningState(true)
		redismgr.restarting = false

		redismgr.client.ScriptFlush()
		for path, _ := range redismgr.scriptCache {
			redismgr.LoadScript(path)
			zed.ZLog("111111111111 RedisMgr aLoadScript: %s  ", path)
		}

		zed.NewCoroutine(func() {
			redismgr.startHeartBeat()
		})

		zed.ZLog("RedisMgr add； %s database: %d Start---->>>", redismgr.addr, redismgr.database)
	}

	return true
}

//DBAction ...
func (redismgr *RedisMgr) DBAction(cb func(*redis.Client) bool) bool {
	defer func() {
		if err := recover(); err != nil {
			zed.LogStackInfo()
			redismgr.Restart()
		}
	}()

	client := redismgr.client
	if client != nil {

		return cb(client)
	} else {

		return false
	}

	return true
}

//Restart ...
func (redismgr *RedisMgr) Restart() {
	//zed.ZLog("enter Restart function ...")
	redismgr.Lock()
	defer redismgr.Unlock()

	if !redismgr.restarting {
		zed.NewCoroutine(func() {
			redismgr.restarting = true

			redismgr.Stop()
			redismgr.Start()
		})
	}
}

//Stop ...
func (redismgr *RedisMgr) Stop() {
	zed.ZLog("RedisMgr Stop 000 ...")
	redismgr.SetRunningState(false)

	redismgr.Lock()
	defer redismgr.Unlock()
	//zed.ZLog("return back Stop function ...")

	err := redismgr.client.Close()
	if err != nil {
		zed.ZLog("RedisMgr stop client(%#v) err: %v", redismgr.client, err)
		panic(err)
	}
}

//startHeartBeat ...
func (redismgr *RedisMgr) startHeartBeat() {
	for {
		select {
		case _, ok := <-redismgr.ticker.C:
			if ok {
				redismgr.heartBeat()
			} else {

				break
			}
		}
	}
}

//heartBeat ...
func (redismgr *RedisMgr) heartBeat() {
	/*
		redismgr.Lock()
		defer redismgr.Unlock()
	*/
	if redismgr.running {
		_, err := redismgr.client.Ping().Result()
		if err != nil {
			zed.ZLog("RedisMgr heartBeat err: %v!", err)
			//panic(err)
			//redismgr.Unlock()
			redismgr.Restart()
		}
	}
}

//IsRuning ...
func (redismgr *RedisMgr) IsRuning() bool {
	redismgr.Lock()
	defer redismgr.Unlock()

	return redismgr.running
}

func (redismgr *RedisMgr) SetRunningState(b bool) {
	//zed.ZLog("enter SetRunningState function ...")
	redismgr.Lock()
	defer redismgr.Unlock()

	redismgr.running = b
	//ed.ZLog("change running state ...")
}

//GetRedisMgrByName ...
func GetRedisMgrByName(name string) *RedisMgr {
	redismgr, isexist := redisMgrs[name]
	if !isexist {
		zed.ZLog("GetRedisMgrByName Error: %s is not exist!", name)

		return nil
	}

	return redismgr
}

//GetRedisMgrPoolByName ...
func GetRedisMgrPoolByName(name string) (*RedisMgrPool, bool) {
	mgr, isexist := redisMgrPools[name]

	return mgr, isexist
}
