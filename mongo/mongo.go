package mongo

import (
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	"github.com/ccppluagopy/zed"
	"sync"
	"time"
)

const (
	DB_DIAL_TIMEOUT   = time.Second * 10
	DB_DIAL_MAX_TIMES = 1000
)

type MongoActionCB func(mongo *MongoMgr) bool

type MongoMgr struct {
	sync.RWMutex
	Session    *mgo.Session
	Collection *mgo.Collection
	tryCount   int
	addr       string
	database   string
	collection string
	usr        string
	passwd     string
	ticker     *time.Ticker
	running    bool
	restarting bool
}

type MongoMgrPool struct {
	mgrs []*MongoMgr
}

var (
	mongoMgrs     = make(map[string]*MongoMgr)
	mongoMgrPools = make(map[string]*MongoMgrPool)
)

func (pool *MongoMgrPool) GetMgr(idx int) *MongoMgr {
	//Println("-----------GetMgr: ", idx, len(pool.mgrs))
	return pool.mgrs[idx%len(pool.mgrs)]
}

func (pool *MongoMgrPool) DBAction(idx int, cb func(*mgo.Collection) bool) bool {
	return pool.GetMgr(idx).DBAction(cb)
}

func (pool *MongoMgrPool) Collection(idx int) *mgo.Collection {
	return pool.GetMgr(idx).Collection
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
	/*zed.ZLog("MongoMgr startHeartbeat addr: %s dbname: %s collection: %s", mongoMgr.addr, mongoMgr.database, mongoMgr.collection)*/
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
			zed.ZLog("MongoMgr Start err: %v addr: %s dbname: %s collection: %s", err, mongoMgr.addr, mongoMgr.database, mongoMgr.collection)
			if mongoMgr.tryCount < DB_DIAL_MAX_TIMES {
				mongoMgr.tryCount = mongoMgr.tryCount + 1
				time.Sleep(time.Second * 1)
				return mongoMgr.Start()
			} else {
				return false
			}
		}

		mongoMgr.Session = session
		mongoMgr.Collection = session.DB(mongoMgr.database).C(mongoMgr.collection)
		//mongoMgr.Session.SetMode(mgo.Monotonic, true)
		//mongoMgr.DB = mongoMgr.Session.DB(mongoMgr.dbname)

		mongoMgr.tryCount = 0

		mongoMgr.ticker = time.NewTicker(time.Hour)

		mongoMgr.SetRunningState(true)
		mongoMgr.restarting = false

		zed.NewCoroutine(func() {
			//mongoMgr.chAction = make(chan MongoActionCB)
			mongoMgr.startHeartbeat()
		})
		/*zed.ZLog("MongoMgr startHeartbeat addr: %s dbname: %s collection: %s", mongoMgr.addr, mongoMgr.database, mongoMgr.collection)*/

		zed.ZLog("MongoMgr addr: %s dbname: %s collection: %s Start() --->>>", mongoMgr.addr, mongoMgr.database, mongoMgr.collection)
	}

	return true
}

func (mongoMgr *MongoMgr) Restart() {
	mongoMgr.Lock()
	defer mongoMgr.Unlock()

	if !mongoMgr.restarting {
		mongoMgr.restarting = true
		zed.NewCoroutine(func() {
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

/*func (mongoMgr *MongoMgr) Collection() *mgo.Collection {
	return mongoMgr.Collection
}*/

func (mongoMgr *MongoMgr) DBAction(cb func(*mgo.Collection) bool) bool {
	defer func() {
		if err := recover(); err != nil {
			/*zed.ZLog("MongoMgr DBAction err: %v!", err)*/
			zed.LogStackInfo()
			mongoMgr.Restart()
		}
	}()

	c := mongoMgr.Collection
	if c != nil {
		return cb(c)
	} else {
		return false
	}

	return true
}

func (mongoMgr *MongoMgr) heartbeat() {
	mongoMgr.RLock()
	defer mongoMgr.RUnlock()

	if mongoMgr.Session != nil {
		if err := mongoMgr.Session.Ping(); err != nil {
			zed.ZLog("MongoMgr heartbeat err: %v!", err)
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
			zed.ZLog("NewMongoMgr %s mgr.Start() Error.!", name)
			return nil
		}

		return mgr
	} else {
		zed.ZLog("NewMongoMgr Error: %s has been exist!", name)
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
			zed.ZLog("NewMongoMgr %s mgr.Start() Error.!", name)
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
				zed.ZLog("%s mgrCopy.Start() %d Error.!", name, i)
				return nil
			}
			mgrs.mgrs[i] = mgrCopy
			//Println("mongo copy", i, mgr.Session, mgrs.mgrs[i].Session)
		}

		mongoMgrPools[name] = mgrs

		return mgrs
	} else {
		zed.ZLog("NewMongoMgrPool Error: %s has been exist!", name)
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
