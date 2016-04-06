package zed

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
)

var (
	mysqlMgrs = make(map[string]*MongoMgr)
)

type MysqlMgr struct {
	sync.RWMutex
	DB       *sql.DB
	tryCount int
	errCB    DBErrorHandler
}

func (msqlMgr *MysqlMgr) Start(addr string, dbname string, usr string, passwd string) bool {
	var err error
	msqlMgr.DB, err = sql.Open("mysql", "user:password@/dbname")
	if err != nil && msqlMgr.tryCount < DB_DIAL_MAX_TIMES {
		msqlMgr.tryCount = msqlMgr.tryCount + 1
		return msqlMgr.Start(addr, dbname, usr, passwd)
	}

	msqlMgr.tryCount = 0

	go msqlMgr.heartbeat()

	return true
}

func (msqlMgr *MysqlMgr) heartbeat() {
	for {
		time.Sleep(time.Hour)
		if msqlMgr.DB != nil {
			ZLog("msql mgr heartbeat!")
			if err := msqlMgr.DB.Ping(); err != nil {
				ZLog("msql mgr heartbeat err %v!", err)
			}
		} else {
			break
		}
	}
}

func (msqlMgr *MysqlMgr) Stop() {
	msqlMgr.Lock()
	defer msqlMgr.Unlock()

	if msqlMgr.DB != nil {
		msqlMgr.DB.Close()
		msqlMgr.DB = nil
	}
}

func NewMysqlMgr(name string, addr string, dbname string, usr string, passwd string, cb DBErrorHandler) *MysqlMgr {
	mgr, ok := mysqlMgrs[name]
	if !ok {
		mgr = &MysqlMgr{
			tryCount: 0,
			errCB:    cb,
		}

		return mgr
	} else {
		ZLog("NewMysqlMgr Error: %s has been exist!", name)
	}

	return mgr
}

func GetMysqlMgrByName(name string) (*MongoMgr, bool) {
	mgr, ok := mysqlMgrs[name]
	return mgr, ok
}
