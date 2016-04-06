package zed

import (
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	//"time"
	"sync"
)

var (
	mongoMgrs = make(map[string]*MongoMgr)
)

type MongoMgr struct {
	sync.RWMutex
	Session  *mgo.Session
	DB       *mgo.Database
	tryCount int
	errCB    DBErrorHandler
}

func (mongoMgr *MongoMgr) Start(addr string, dbname string, usr string, passwd string) bool {
	var err error

	// 3 new goroutines created when mgo.Dial success
	mongoMgr.Session, err = mgo.DialWithTimeout(addr, DB_DIAL_TIMEOUT)
	if err != nil && mongoMgr.tryCount < DB_DIAL_MAX_TIMES {
		mongoMgr.tryCount = mongoMgr.tryCount + 1
		return mongoMgr.Start(addr, dbname, usr, passwd)
	}

	mongoMgr.Session.SetMode(mgo.Monotonic, true)
	mongoMgr.DB = mongoMgr.Session.DB(dbname)

	mongoMgr.StuInfo = mongoMgr.DB.C("StuInfo")

	mongoMgr.tryCount = 0

	return true
}

func (mongoMgr *MongoMgr) Stop() {
	mongoMgr.Lock()
	defer mongoMgr.Unlock()

	if mongoMgr.Session != nil {
		mongoMgr.Session.Close()
		mongoMgr.Session = nil
	}
}

func NewMongoMgr(name string, addr string, dbname string, usr string, passwd string, cb DBErrorHandler) *MongoMgr {
	mgr, ok := mongoMgrs[name]
	if !ok {
		mgr = &MongoMgr{
			tryCount: 0,
			errCB:    cb,
		}

		return mgr
	} else {
		ZLog("NewMongoMgr Error: %s has been exist!", name)
	}

	return mgr
}

func GetMongoMgrByName(name string) (*MongoMgr, bool) {
	mgr, ok := mongoMgrs[name]
	return mgr, ok
}
