package zed

import (
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

var (
	mongoMgrs = make(map[string]*MongoMgr)
)

type MongoActionCB func(mongo *MongoMgr) bool

type MongoMgr struct {
	sync.RWMutex
	Session    *mgo.Session
	tryCount   int
	addr       string
	database   string
	collection string
	usr        string
	passwd     string
	//chAction chan MongoActionCB
	ticker  *time.Ticker
	running bool
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
	/*var (
		cb       MongoActionCB
		ok       bool
		chAction = mongoMgr.chAction
	)*/
	Printf("MongoMgr start heartbeat \n")
	for {
		select {
		/*case cb, ok = <-chAction:
		if ok && mongoMgr.IsRunning() {
			if !cb(mongoMgr) {
				NewCoroutine(func() {
					mongoMgr.Restart()
				})
			}
		} else {
			break
		}*/
		case <-mongoMgr.ticker.C:
			mongoMgr.heartbeat()
		}
	}
}

func (mongoMgr *MongoMgr) Start() {
	var err error

	if !mongoMgr.IsRunning() {
		mongoMgr.Session, err = mgo.DialWithTimeout(mongoMgr.addr, DB_DIAL_TIMEOUT)
		if err != nil {
			Printf("MongoMgr Start err: %v .............\n", err)
			if mongoMgr.tryCount < DB_DIAL_MAX_TIMES {
				mongoMgr.tryCount = mongoMgr.tryCount + 1
				mongoMgr.Start()

				return
			} else {
				return
			}
		}

		//mongoMgr.Session.SetMode(mgo.Monotonic, true)
		//mongoMgr.DB = mongoMgr.Session.DB(mongoMgr.dbname)

		mongoMgr.tryCount = 0

		mongoMgr.ticker = time.NewTicker(time.Hour)

		mongoMgr.SetRunningState(true)

		NewCoroutine(func() {
			//mongoMgr.chAction = make(chan MongoActionCB)
			mongoMgr.startHeartbeat()
		})

		Printf("MongoMgr addr: %s dbname: %s Start() --->>>\n", mongoMgr.addr, mongoMgr.database)
	}
}

func (mongoMgr *MongoMgr) Restart() {
	NewCoroutine(func() {
		mongoMgr.Stop()
		mongoMgr.Start()
	})
}

func (mongoMgr *MongoMgr) Stop() {
	mongoMgr.Lock()
	defer mongoMgr.Unlock()

	if mongoMgr.running {
		mongoMgr.running = false
		if mongoMgr.Session != nil {
			mongoMgr.Session.Close()
			mongoMgr.Session = nil
		}
		/*if mongoMgr.chAction != nil {
			close(mongoMgr.chAction)
		}*/
	}
}

func (mongoMgr *MongoMgr) DBAction(cb func(*mgo.Collection)) {
	mongoMgr.RLock()
	defer func() {
		mongoMgr.RUnlock()
		if err := recover(); err != nil {
			LogError(LOG_IDX, LOG_IDX, "MongoMgr DBAction err: %v!", err)
			mongoMgr.Restart()
		}
	}()

	if mongoMgr.running {
		session := mongoMgr.Session.Clone()
		defer func() {
			session.Close()
			if err := recover(); err != nil {
				panic(err)
			}
		}()
		//c := session.DB(mongoMgr.database).C(mongoMgr.collection)
		//cb(c)
		cb(session.DB(mongoMgr.database).C(mongoMgr.collection))
	} else {
		cb(nil)
	}
}

func (mongoMgr *MongoMgr) heartbeat() {
	mongoMgr.RLock()
	defer mongoMgr.RUnlock()

	if mongoMgr.Session != nil {
		if err := mongoMgr.Session.Ping(); err != nil {
			LogError(LOG_IDX, LOG_IDX, "MongoMgr heartbeat err: %v!", err)
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
			running: false,
		}
		mgr.Start()
		return mgr
	} else {
		LogError(LOG_IDX, LOG_IDX, "NewMongoMgr Error: %s has been exist!", name)
	}

	return mgr
}

func GetMongoMgrByName(name string) (*MongoMgr, bool) {
	mgr, ok := mongoMgrs[name]
	return mgr, ok
}
