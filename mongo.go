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

type MongoMgr struct {
	sync.RWMutex
	Session  *mgo.Session
	DB       *mgo.Database
	tryCount int
	addr     string
	dbname   string
	usr      string
	passwd   string
	chAction chan MongoActionCB
	ticker   *time.Ticker
	running  bool
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

func (mongoMgr *MongoMgr) handleAction() {
	var (
		cb       MongoActionCB
		ok       bool
		chAction = mongoMgr.chAction
	)

	select {
	case cb, ok = <-chAction:
		if ok && mongoMgr.IsRunning() {
			if !cb(mongoMgr) {
				NewCoroutine(func() {
					mongoMgr.Restart()
				})
			}
		} else {
			break
		}
	case <-mongoMgr.ticker.C:
		mongoMgr.heartbeat()
	}
}

func (mongoMgr *MongoMgr) Start() {
	var err error

	if !mongoMgr.IsRunning() {
		mongoMgr.Session, err = mgo.DialWithTimeout(mongoMgr.addr, DB_DIAL_TIMEOUT)
		if err != nil {
			if mongoMgr.tryCount < DB_DIAL_MAX_TIMES {
				mongoMgr.tryCount = mongoMgr.tryCount + 1
				mongoMgr.Start()

				return
			} else {
				return
			}
		}

		mongoMgr.Session.SetMode(mgo.Monotonic, true)
		mongoMgr.DB = mongoMgr.Session.DB(mongoMgr.dbname)

		mongoMgr.tryCount = 0

		mongoMgr.ticker = time.NewTicker(time.Hour)

		mongoMgr.SetRunningState(true)

		NewCoroutine(func() {
			mongoMgr.chAction = make(chan MongoActionCB)
			mongoMgr.handleAction()
		})
	}
}

func (mongoMgr *MongoMgr) Restart() {
	mongoMgr.Stop()
	mongoMgr.Start()
}

func (mongoMgr *MongoMgr) Stop() {
	mongoMgr.Lock()
	defer mongoMgr.Unlock()

	if mongoMgr.running {
		mongoMgr.running = false
		if mongoMgr.Session != nil {
			mongoMgr.Session.Close()
			mongoMgr.Session = nil
			mongoMgr.DB = nil
		}
		if mongoMgr.chAction != nil {
			close(mongoMgr.chAction)
		}
	}
}

func (mongoMgr *MongoMgr) DBAction(cb MongoActionCB) {
	mongoMgr.Lock()
	defer mongoMgr.Unlock()

	if mongoMgr.running {
		mongoMgr.chAction <- cb
	} else {
		cb(nil)
	}
}

func (mongoMgr *MongoMgr) heartbeat() {
	mongoMgr.DBAction(func(mongo *MongoMgr) bool {
		if mongo.Session != nil {
			if err := mongo.Session.Ping(); err != nil {
				LogError(LOG_IDX, LOG_IDX, "MongoMgr heartbeat err: %v!", err)
				return false
			}
		} else {

		}
		return true
	})
}

func NewMongoMgr(name string, addr string, dbname string, usr string, passwd string) *MongoMgr {
	mgr, ok := mongoMgrs[name]
	if !ok {
		mgr = &MongoMgr{
			tryCount: 0,
			Session:  nil,
			DB:       nil,
			addr:     addr,
			dbname:   dbname,
			usr:      usr,
			passwd:   passwd,
			chAction: nil,
			running:  false,
		}

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
