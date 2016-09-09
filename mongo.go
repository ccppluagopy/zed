package zed

import (
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	//"sync"
	"time"
)

type MongoMgrs []*MongoMgr

var (
	mongoMgrs     = make(map[string]*MongoMgr)
	mongoMgrPools = make(map[string]*MongoMgrPool)
)

func (pool *MongoMgrPool) GetMgr(idx int) *MongoMgr {
	//Println("-----------GetMgr: ", idx, len(pool.mgrs))
	return pool.mgrs[idx%len(pool.mgrs)]
}

func (pool *MongoMgrPool) DBAction(idx int, cb func(*mgo.Collection)) {
	pool.GetMgr(idx).DBAction(cb)
}

func (pool *MongoMgrPool) Collection(idx int) *mgo.Collection {
	return pool.GetMgr(idx).Collection()
}

func (pool *MongoMgrPool) Stop() {
	for i := 0; i < len(pool.mgrs); i++ {
		pool.mgrs[i].Stop()
	}
}

func (mongoMgr *MongoMgr) IsRunning() bool {
	mongoMgr.Lock()
	defer mongoMgr.Unlock()

	return mongoMgr.running
}

func (mongoMgr *MongoMgr) SetRunningState(running bool) {
	mongoMgr.Lock()
	defer mongoMgr.Unlock()

	mongoMgr.running = running
}

func (mongoMgr *MongoMgr) startHeartbeat() {
	ZLog("MongoMgr start heartbeat")
	for {
		select {
		case _, ok := <-mongoMgr.ticker.C:
			if ok {
				mongoMgr.heartbeat()
			} else {
				break
			}
		}
	}
}

func (mongoMgr *MongoMgr) Start() bool {
	if !mongoMgr.IsRunning() {
		session, err := mgo.DialWithTimeout(mongoMgr.addr, DB_DIAL_TIMEOUT)
		if err != nil {
			ZLog("MongoMgr Start err: %v .............", err)
			if mongoMgr.tryCount < DB_DIAL_MAX_TIMES {
				mongoMgr.tryCount = mongoMgr.tryCount + 1

				return mongoMgr.Start()
			} else {
				return false
			}
		}

		mongoMgr.Session = session
		//mongoMgr.Session.SetMode(mgo.Monotonic, true)
		//mongoMgr.DB = mongoMgr.Session.DB(mongoMgr.dbname)

		mongoMgr.tryCount = 0

		mongoMgr.ticker = time.NewTicker(time.Hour)

		mongoMgr.SetRunningState(true)
		mongoMgr.restarting = false

		NewCoroutine(func() {
			//mongoMgr.chAction = make(chan MongoActionCB)
			mongoMgr.startHeartbeat()
		})

		ZLog("MongoMgr addr: %s dbname: %s Start() --->>>", mongoMgr.addr, mongoMgr.database)
	}

	return true
}

func (mongoMgr *MongoMgr) Restart() {
	mongoMgr.Lock()
	defer mongoMgr.Unlock()

	if !mongoMgr.restarting {
		mongoMgr.restarting = true
		NewCoroutine(func() {
			mongoMgr.Stop()
			mongoMgr.Start()
		})
	}
}

func (mongoMgr *MongoMgr) Stop() {
	mongoMgr.Lock()
	defer mongoMgr.Unlock()
	if mongoMgr.running {
		mongoMgr.running = false
		mongoMgr.ticker.Stop()
		if mongoMgr.Session != nil {
			mongoMgr.Session.Close()
			mongoMgr.Session = nil
		}
		/*if mongoMgr.chAction != nil {
			close(mongoMgr.chAction)
		}*/
	}
}

func (mongoMgr *MongoMgr) Collection() *mgo.Collection {
	return mongoMgr.Session.DB(mongoMgr.database).C(mongoMgr.collection)
}

func (mongoMgr *MongoMgr) DBAction(cb func(*mgo.Collection)) {
	/*mongoMgr.RLock()
	defer mongoMgr.Unlock()
	*/
	//if mongoMgr.running {
	//Println("DBAction ....")
	defer func() {
		if err := recover(); err != nil {
			LogError(LOG_IDX, LOG_IDX, "MongoMgr DBAction err: %v!", err)
			mongoMgr.Restart()
		}
	}()
	session := mongoMgr.Session //.Clone()
	cb(session.DB(mongoMgr.database).C(mongoMgr.collection))
	/*} else {
		cb(nil)
	}*/
}

func (mongoMgr *MongoMgr) heartbeat() {
	mongoMgr.RLock()
	defer mongoMgr.RUnlock()

	if mongoMgr.Session != nil {
		if err := mongoMgr.Session.Ping(); err != nil {
			LogError(LOG_IDX, LOG_IDX, "MongoMgr heartbeat err: %v!", err)
			panic(err)
			mongoMgr.Restart()
		}
	} else {

	}
}

func NewMongoMgr(name string, addr string, dbname string, cname string, usr string, passwd string) *MongoMgr {
	mgr, ok := mongoMgrs[name]
	if !ok {
		mgr = &MongoMgr{
			tryCount:   0,
			Session:    nil,
			addr:       addr,
			database:   dbname,
			collection: cname,
			usr:        usr,
			passwd:     passwd,
			//chAction: nil,
			running:    false,
			restarting: false,
		}
		ok := mgr.Start()
		if !ok {
			LogError(LOG_IDX, LOG_IDX, "NewMongoMgr %s mgr.Start() Error.!", name)
			return nil
		}

		return mgr
	} else {
		LogError(LOG_IDX, LOG_IDX, "NewMongoMgr Error: %s has been exist!", name)
	}

	return nil
}

func NewMongoMgrPool(name string, addr string, dbname string, cname string, usr string, passwd string, size int) *MongoMgrPool {
	mgrs, ok := mongoMgrPools[name]
	if !ok {
		mgrs = &MongoMgrPool{
			mgrs: make([]*MongoMgr, size),
		}
		mgr := &MongoMgr{
			tryCount:   0,
			Session:    nil,
			addr:       addr,
			database:   dbname,
			collection: cname,
			usr:        usr,
			passwd:     passwd,
			//chAction: nil,
			running:    false,
			restarting: false,
		}
		ok := mgr.Start()
		if !ok {
			LogError(LOG_IDX, LOG_IDX, "NewMongoMgr %s mgr.Start() Error.!", name)
			return nil
		}

		mgrs.mgrs[0] = mgr
		for i := 1; i < size; i++ {
			mgrCopy := &MongoMgr{
				tryCount:   0,
				Session:    mgr.Session.Clone(),
				addr:       addr,
				database:   dbname,
				collection: cname,
				usr:        usr,
				passwd:     passwd,
				//chAction: nil,
				running:    false,
				restarting: false,
			}
			ok := mgrCopy.Start()
			if !ok {
				LogError(LOG_IDX, LOG_IDX, "%s mgrCopy.Start() %d Error.!", name, i)
				return nil
			}
			mgrs.mgrs[i] = mgrCopy
			//Println("mongo copy", i, mgr.Session, mgrs.mgrs[i].Session)
		}

		mongoMgrPools[name] = mgrs

		return mgrs
	} else {
		LogError(LOG_IDX, LOG_IDX, "NewMongoMgrPool Error: %s has been exist!", name)
	}

	return nil
}

func GetMongoMgrByName(name string) (*MongoMgr, bool) {
	mgr, ok := mongoMgrs[name]
	return mgr, ok
}

func GetMongoMgrPoolByName(name string) (*MongoMgrPool, bool) {
	mgr, ok := mongoMgrPools[name]
	return mgr, ok
}
