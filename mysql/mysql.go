package mysql

import (
	/*"database/sql"
	"github.com/go-sql-driver/mysql"*/
	"github.com/ccppluagopy/zed"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"sync"
	"time"
)

var (
	mysqlMgrs     = make(map[string]*MysqlMgr)
	mysqlMgrPools = make(map[string]*MysqlMgrPool)
)

const (
	DB_DIAL_TIMEOUT   = time.Second * 10
	DB_DIAL_MAX_TIMES = 1000
)

type MysqlActionCB func(*mysql.Conn)

type MysqlMgr struct {
	sync.RWMutex
	DB *mysql.Conn
	//DB       *sql.DB
	tryCount   int
	addr       string
	dbname     string
	usr        string
	passwd     string
	ticker     *time.Ticker
	running    bool
	restarting bool
}

type MysqlMgrPool struct {
	mgrs []*MysqlMgr
}

func (pool *MysqlMgrPool) GetMgr(idx int) *MysqlMgr {
	//Println("MysqlMgrPool GetMgr: ", idx, len(pool.mgrs))
	return pool.mgrs[idx%len(pool.mgrs)]
}

func (pool *MysqlMgrPool) DBAction(idx int, cb func(*mysql.Conn)) {
	(*(pool.GetMgr(idx))).DBAction(cb)
}

func (pool *MysqlMgrPool) DB(idx int) *mysql.Conn {
	return (pool.GetMgr(idx).DB)
}

func (pool *MysqlMgrPool) Stop() {
	for i := 0; i < len(pool.mgrs); i++ {
		pool.mgrs[i].Stop()
	}
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

/*
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
}*/

func (mysqlMgr *MysqlMgr) startHeartbeat() {
	zed.ZLog("MysqlMgr start heartbeat")
	for {
		select {
		case _, ok := <-mysqlMgr.ticker.C:
			if ok {
				mysqlMgr.heartbeat()
			} else {
				break
			}
		}
	}
}

func (msqlMgr *MysqlMgr) Start() bool {
	var err error

	if !msqlMgr.IsRunning() {
		if msqlMgr.restarting {
			err = (*(msqlMgr.DB)).Reconnect()
		} else {
			err = (*(msqlMgr.DB)).Connect()
		}

		//msqlMgr.DB, err = msqlMgr.Session.Use()
		if err != nil {
			if msqlMgr.tryCount < DB_DIAL_MAX_TIMES {
				msqlMgr.tryCount = msqlMgr.tryCount + 1
				time.Sleep(time.Second * 1)
				return msqlMgr.Start()
			} else {
				return false
			}
		}

		/*msqlMgr.Session.SetMode(mgo.Monotonic, true)
		msqlMgr.DB = msqlMgr.Session.DB(msqlMgr.dbname)*/

		msqlMgr.tryCount = 0

		msqlMgr.ticker = time.NewTicker(time.Hour)

		msqlMgr.SetRunningState(true)
		msqlMgr.restarting = false

		zed.NewCoroutine(func() {
			/*msqlMgr.chAction = make(chan MysqlActionCB)
			msqlMgr.handleAction()*/
			msqlMgr.startHeartbeat()
		})

		zed.ZLog("MsqlMgr addr: %s dbname: %v Start() --->>>", msqlMgr.addr, msqlMgr.DB)
	}
	//Println("-----------------------------------")
	return true
}

func (msqlMgr *MysqlMgr) Restart() {
	zed.NewCoroutine(func() {
		msqlMgr.RLock()
		defer msqlMgr.RUnlock()

		if !msqlMgr.restarting {
			msqlMgr.restarting = true
			zed.NewCoroutine(func() {
				msqlMgr.Stop()
				msqlMgr.Start()
			})
		}
	})
}

func (msqlMgr *MysqlMgr) Stop() {
	msqlMgr.Lock()
	defer msqlMgr.Unlock()
	if msqlMgr.running {
		msqlMgr.running = false
		msqlMgr.ticker.Stop()
		if msqlMgr.DB != nil {
			//(*(msqlMgr.DB)).Close()
			//msqlMgr.DB = nil
		}
	}
}

func (msqlMgr *MysqlMgr) DBAction(cb func(*mysql.Conn)) {
	msqlMgr.Lock()

	defer func() {
		msqlMgr.Unlock()
		if err := recover(); err != nil {
			/*zed.ZLog("MongoMgr DBAction err: %v!", err)*/
			zed.LogStackInfo()
			mongoMgr.Restart()
		}
	}()

	//db := msqlMgr.DB
	if msqlMgr.running {
		cb(msqlMgr.DB)
	}
}

func (msqlMgr *MysqlMgr) heartbeat() {
	msqlMgr.Lock()
	defer msqlMgr.Unlock()

	msqlMgr.DBAction(func(msql *mysql.Conn) {
		if msql != nil {
			if err := (*msql).Ping(); err != nil {
				zed.LogError(zed.LOG_IDX, zed.LOG_IDX, "MysqlMgr heartbeat err: %v!", err)
				panic(err)
			}
		} else {

		}
		return
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
			//chAction: nil,
			ticker:  nil,
			running: false,
		}

		m := mysql.New("tcp", "", addr, usr, passwd, dbname)
		mgr.DB = &m
		ok := mgr.Start()
		if !ok {
			zed.LogError(zed.LOG_IDX, zed.LOG_IDX, "NewMysqlMgr %s mgr.Start() Error.!", name)
			return nil
		}

		return mgr
	} else {
		zed.LogError(zed.LOG_IDX, zed.LOG_IDX, "NewMysqlMgr Error: %s has been exist!", name)
	}

	return mgr
}

func GetMysqlMgrByName(name string) (*MysqlMgr, bool) {
	mgr, ok := mysqlMgrs[name]
	return mgr, ok
}

func NewMysqlMgrPool(name string, addr string, dbname string, usr string, passwd string, size int) *MysqlMgrPool {
	mgrs, ok := mysqlMgrPools[name]
	if !ok {
		mgrs = &MysqlMgrPool{
			mgrs: make([]*MysqlMgr, size),
		}

		mgr := &MysqlMgr{
			tryCount: 0,
			DB:       nil,
			//DB:       nil,
			addr:   addr,
			dbname: dbname,
			usr:    usr,
			passwd: passwd,
			//chAction: nil,
			ticker:     nil,
			running:    false,
			restarting: false,
		}
		m := mysql.New("tcp", "", addr, usr, passwd, dbname)
		mgr.DB = &m
		ok := mgr.Start()
		if !ok {
			zed.LogError(zed.LOG_IDX, zed.LOG_IDX, "NewMysqlMgr %s mgr.Start() Error.!", name)
			return nil
		}

		mgrs.mgrs[0] = mgr

		for i := 1; i < size; i++ {
			mgrCopy := &MysqlMgr{
				tryCount: 0,
				DB:       nil,
				//DB:       nil,
				addr:   addr,
				dbname: dbname,
				usr:    usr,
				passwd: passwd,
				//chAction: nil,
				ticker:     nil,
				running:    false,
				restarting: false,
			}
			m2 := m.Clone()
			mgrCopy.DB = &m2

			ok := mgrCopy.Start()
			if !ok {
				zed.LogError(zed.LOG_IDX, zed.LOG_IDX, "%s mgrCopy.Start() %d Error.!", name, i)
				return nil
			}
			mgrs.mgrs[i] = mgrCopy
			//Println("mongo copy", i, mgr.Session, mgrs.mgrs[i].Session)
		}

		mysqlMgrPools[name] = mgrs

		return mgrs
	} else {
		zed.LogError(zed.LOG_IDX, zed.LOG_IDX, "NewMysqlMgrPool Error: %s has been exist!", name)
	}

	return nil
}

func GetMysqlMgrPoolByName(name string) (*MysqlMgrPool, bool) {
	mgr, ok := mysqlMgrPools[name]
	return mgr, ok
}
