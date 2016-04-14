package zed

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"time"
)

var (
	mysqlMgrs = make(map[string]*MysqlMgr)
)

type MysqlMgr struct {
	sync.RWMutex
	//Session  *mgo.Session
	DB       *sql.DB
	tryCount int
	addr     string
	dbname   string
	usr      string
	passwd   string
	chAction chan MysqlActionCB
	ticker   *time.Ticker
	running  bool
}

func (msqlMgr *MysqlMgr) IsRunning() bool {
	msqlMgr.Lock()
	defer msqlMgr.Unlock()

	return msqlMgr.running
}

func (msqlMgr *MysqlMgr) SetRunningState(running bool) {
	msqlMgr.Lock()
	defer msqlMgr.Unlock()

	msqlMgr.running = running
}

func (msqlMgr *MysqlMgr) handleAction() {
	var (
		cb       MysqlActionCB
		ok       bool
		chAction = msqlMgr.chAction
	)

	for {
		select {
		case cb, ok = <-chAction:
			if ok && msqlMgr.IsRunning() {
				if !cb(msqlMgr) {
					NewCoroutine(func() {
						msqlMgr.Restart()
					})
				}
			} else {
				break
			}
		case <-msqlMgr.ticker.C:
			msqlMgr.heartbeat()
		}
	}
}

func (msqlMgr *MysqlMgr) Start() {
	var err error

	if !msqlMgr.IsRunning() {
		//msqlMgr.Session, err = mgo.DialWithTimeout(msqlMgr.addr, DB_DIAL_TIMEOUT)
		msqlMgr.DB, err = sql.Open("mysql", "user:password@/dbname")
		if err != nil {
			if msqlMgr.tryCount < DB_DIAL_MAX_TIMES {
				msqlMgr.tryCount = msqlMgr.tryCount + 1
				msqlMgr.Start()

				return
			} else {
				return
			}
		}

		/*msqlMgr.Session.SetMode(mgo.Monotonic, true)
		msqlMgr.DB = msqlMgr.Session.DB(msqlMgr.dbname)*/

		msqlMgr.tryCount = 0

		msqlMgr.ticker = time.NewTicker(time.Hour)

		msqlMgr.SetRunningState(true)

		NewCoroutine(func() {
			msqlMgr.chAction = make(chan MysqlActionCB)
			msqlMgr.handleAction()
		})
	}
}

func (msqlMgr *MysqlMgr) Restart() {
	msqlMgr.Stop()
	msqlMgr.Start()
}

func (msqlMgr *MysqlMgr) Stop() {
	msqlMgr.Lock()
	defer msqlMgr.Unlock()

	if msqlMgr.running {
		msqlMgr.running = false
		if msqlMgr.DB != nil {
			msqlMgr.DB.Close()
			msqlMgr.DB = nil
		}
		if msqlMgr.chAction != nil {
			close(msqlMgr.chAction)
		}
		if msqlMgr.ticker != nil {
			msqlMgr.ticker.Stop()
		}
	}
}

func (msqlMgr *MysqlMgr) DBAction(cb MysqlActionCB) {
	msqlMgr.Lock()
	defer msqlMgr.Unlock()

	if msqlMgr.running {
		msqlMgr.chAction <- cb
	} else {
		cb(nil)
	}
}

func (msqlMgr *MysqlMgr) heartbeat() {
	msqlMgr.DBAction(func(msql *MysqlMgr) bool {
		if msql.DB != nil {
			if err := msql.DB.Ping(); err != nil {
				LogError(LOG_IDX, LOG_IDX, "MysqlMgr heartbeat err: %v!", err)
				return false
			}
		} else {

		}
		return true
	})
}

func NewMysqlMgr(name string, addr string, dbname string, usr string, passwd string) *MysqlMgr {
	mgr, ok := mysqlMgrs[name]
	if !ok {
		mgr = &MysqlMgr{
			tryCount: 0,
			DB:       nil,
			addr:     addr,
			dbname:   dbname,
			usr:      usr,
			passwd:   passwd,
			chAction: nil,
			ticker:   nil,
			running:  false,
		}

		return mgr
	} else {
		LogError(LOG_IDX, LOG_IDX, "NewMysqlMgr Error: %s has been exist!", name)
	}

	return mgr
}

func GetMysqlMgrByName(name string) (*MysqlMgr, bool) {
	mgr, ok := mysqlMgrs[name]
	return mgr, ok
}
