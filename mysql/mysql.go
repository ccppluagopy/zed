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
	mysqlmtx  = sync.Mutex{}
	instances = make(map[string]*Mysql)
	poolmtx   = sync.Mutex{}
	pools     = make(map[string]*MysqlPool)

	MysqlErr = &MysqlError{}
)

const (
	DB_DIAL_TIMEOUT     = time.Second * 10
	DB_DIAL_MAX_TIMES   = 1000
	KEEP_ALIVE_INTERVAL = time.Hour
)

const (
	STATE_RUNNING = iota
	STATE_RECONNECTING
	STATE_STOP
)

type MysqlError struct {
}

func (err *MysqlError) Error() string {
	return "MysqlError For Reconnect"
}

type Mysql struct {
	sync.RWMutex
	DB     mysql.Conn
	addr   string
	dbname string
	usr    string
	passwd string
	timer  *time.Timer
	state  int
}

func (msql *Mysql) startHeartbeat() {
	if msql.state == STATE_RUNNING {
		if msql.timer == nil {
			msql.timer = time.NewTimer(KEEP_ALIVE_INTERVAL)
			zed.Async(func() {
				for {
					_, ok := <-msql.timer.C
					if !ok {
						break
					}
					msql.Ping()
					msql.timer.Reset(KEEP_ALIVE_INTERVAL)
				}
			})
		}
	}
}

func (msql *Mysql) Start() {
	//zed.Println("--- Start() 111")
	if msql.state == STATE_STOP {
		//zed.Println("--- Start() 222")
		msql.state = STATE_RECONNECTING
		//msql.Reset()
		zed.Async(msql.Connect)
	}
}

func (msql *Mysql) Stop() {
	msql.Lock()
	defer msql.Unlock()
	if msql.state != STATE_STOP {
		msql.state = STATE_STOP
		msql.Reset()
	}
}

func (msql *Mysql) Reset() {
	if msql.timer != nil {
		msql.timer.Stop()
		msql.timer = nil
	}
	if msql.DB != nil {
		msql.DB.Close()
		msql.DB = nil
	}
}

func (msql *Mysql) Connect() {
	//zed.Println("--- Connect() 111")
	msql.Lock()
	defer msql.Unlock()

	//zed.Println("--- Connect() 222")
	if msql.state != STATE_STOP {
		msql.state = STATE_RECONNECTING
		msql.Reset()
		//zed.Println("--- Connect() 333")
		db := mysql.New("tcp", "", msql.addr, msql.usr, msql.passwd, msql.dbname)
		if err := db.Connect(); err != nil {
			zed.ZLog("Mysql Connect To %s Failed, Error: %s", msql.addr, err.Error())
			zed.Async(func() {
				//zed.Println("--- Connect() 666")
				time.Sleep(time.Second)
				msql.Connect()
			})
			return
		}
		//zed.Println("--- Connect() 444")
		msql.DB = db
		msql.state = STATE_RUNNING
		msql.startHeartbeat()
		zed.ZLog("Mysql Connect Addr: %s DBName: %s", msql.addr, msql.dbname)
	}
}

func (msql *Mysql) DBAction(cb func(mysql.Conn)) {
	msql.Lock()
	defer msql.Unlock()
	//if msql.state == STATE_RUNNING {
	defer func() {
		if err := recover(); err != nil {
			_, ok := err.(*MysqlError)
			if ok && msql.state == STATE_RUNNING {
				zed.ZLog("DBAction Mysql Error")
				zed.Async(func() {
					msql.Connect()
				})
			} else {
				zed.ZLog("DBAction Error: %v", err)
				zed.LogStackInfo()
			}
		} else {
			if msql.timer != nil {
				msql.timer.Reset(KEEP_ALIVE_INTERVAL)
			}
		}
	}()
	cb(msql.DB)
	//}
}

func (msql *Mysql) Ping() {
	msql.DBAction(func(conn mysql.Conn) {
		conn.Ping()
	})
}

func NewMysql(name string, addr string, dbname string, usr string, passwd string) *Mysql {
	mysqlmtx.Lock()
	defer mysqlmtx.Unlock()
	msql, ok := instances[name]
	if !ok {
		msql = &Mysql{
			DB:     nil,
			addr:   addr,
			dbname: dbname,
			usr:    usr,
			passwd: passwd,
			timer:  nil,
			state:  STATE_STOP,
		}

		//msql.DB = mysql.New("tcp", "", addr, usr, passwd, dbname)

		msql.Start()

		return msql
	} else {
		zed.ZLog("NewMysql Error: %s Exist!", name)
	}

	return nil
}

func GetMysqlByName(name string) (*Mysql, bool) {
	msql, ok := instances[name]
	return msql, ok
}

type MysqlPool struct {
	size      int
	instances []*Mysql
}

func (pool *MysqlPool) GetMysql(idx int) *Mysql {
	if pool.size == 0{
		return nil
	}
	return pool.instances[idx%pool.size]
}

func (pool *MysqlPool) DBAction(idx int, cb func(mysql.Conn)) {
	if pool.size == 0{
		cb(nil)
		return
	}
	pool.instances[idx%pool.size].DBAction(cb)
}

func (pool *MysqlPool) GetDB(idx int) mysql.Conn {
	if pool.size == 0{
		return nil
	}
	return pool.instances[idx%pool.size].DB
}

func (pool *MysqlPool) Stop() {
	for i := 0; i < len(pool.instances); i++ {
		pool.instances[i].Stop()
	}
}

func NewMysqlPool(name string, addr string, dbname string, usr string, passwd string, size int) *MysqlPool {
	poolmtx.Lock()
	defer poolmtx.Unlock()
	pool, ok := pools[name]
	if !ok {
		pool = &MysqlPool{size: size}
		for i := 1; i < size; i++ {
			msql := NewMysql(zed.Sprintf("%s_%d", name, i), addr, dbname, usr, passwd)
			pool.instances = append(pool.instances, msql)
		}
		pools[name] = pool
	} else {
		zed.ZLog("NewMysqlPool Error: %s Exist!", name)
	}

	return pool
}

func GetMysqlPoolByName(name string) (*MysqlPool, bool) {
	poolmtx.Lock()
	defer poolmtx.Unlock()
	mgr, ok := pools[name]
	return mgr, ok
}
