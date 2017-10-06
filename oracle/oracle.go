package oracle

import (
	"database/sql"
	"fmt"
	"github.com/naivefox/go-oci8"
	"sync"
	"time"
)

const (
	STATE_RUNNING = iota
	STATE_RECONNECTING
	STATE_STOP
)

const (
	KEEPALIVE_INTERVAL = time.Hour
)

var (
	oracles   = make(map[string]Oracle)
	oraclemtx = sync.Mutex{}

	pools   = make(map[string]*OraclePool)
	poolmtx = sync.Mutex{}

	OrcleErr = &OracleError{}
)

type OracleError struct {
}

func (err *OracleError) Error() string {
	return "naivefox/oracle OracleError"
}

func Recover() {
	panic(OrcleErr)
}

type Oracle struct {
	sync.Mutex
	Name   string
	DB     *sql.DB
	DBInfo string
	ticker *time.Ticker
	state  int
}

func (ora *Oracle) Ping() bool {
	ret := true
	ora.DBAction(func(db *sql.DB) {
		ret = db.Ping() == nil
	})
	return ret
}

func (ora *Oracle) startHeartbeat() {
	if ora.state == STATE_RUNNING {
		if ora.ticker == nil {
			time.AfterFunc(1, func() {
				ora.ticker = ticker.NewTicker(KEEPALIVE_INTERVAL)
				for {
					_, ok := <-ora.ticker.C
					if !ok {
						break
					}
					ora.Ping()
				}
			})
		}
	}
}

func (ora *Oracle) Connect() {
	ora.Lock()
	defer ora.Unlock()

	if ora.state != STATE_STOP {
		ora.state = STATE_RECONNECTING
		ora.Reset()
		db, err := sql.Open("oci8", ora.DBInfo)
		if err != nil {
			fmt.Printf("Oracle Connect To %s Failed, Error: %s\n", ora.DBInfo, err.Error())
			time.AfterFunc(1, func() {
				time.Sleep(time.Second)
				ora.Connect()
			})
			return
		}
		ora.DB = db
		ora.state = STATE_RUNNING
		ora.startHeartbeat()
		fmt.Printf("Oracle(Name: %s, DBInfo: %s) Connected()\n", ora.Name, ora.DBInfo)
	}
}

func (ora *Oracle) Start() {
	if ora.state == STATE_STOP {
		ora.state = STATE_RECONNECTING
		time.AfterFunc(1, ora.Connect())
	}
}

func (ora *Oracle) DBAction(cb func(*sql.DB)) *Oracle {
	ora.Lock()
	defer ora.Unlock()
	defer func() {
		if err := recover(); err != nil {
			_, ok := err.(*OracleError)
			if ok && ora.state == STATE_RUNNING {
				time.AfterFunc(1, ora.Connect)
			}
		} else {
			fmt.Printf("Oracle DBAction Error: %v\n", err)
			//log stacks
		}
	}()
	cb(ora.DB)
	return ora
}

func (ora *Oracle) Reset() {
	if ora.ticker != nil {
		ora.ticker.Stop()
		ora.ticker = nil
	}
	if ora.DB != nil {
		ora.DB.Close()
		ora.DB = nil
	}
}

func (ora *Oracle) Stop() {
	ora.Lock()
	defer ora.Unlock()
	if ora.state != STATE_STOP {
		ora.state = STATE_STOP
		ora.Reset()
	}
}

func (ora *Oracle) NewOracle(name string, dbinfo string) *Oracle {
	if name == "" {
		fmt.Printf("NewOracle Error: name is null\n")
		return nil
	}
	if dbinfo == "" {
		fmt.Printf("NewOracle Error: dbinfo is null\n")
		return nil
	}

	oraclemtx.Lock()
	defer oraclemtx.Unlock()

	ora, ok := oracles[name]
	if !ok {
		ora = &Oracle{
			Name:   name,
			DBInfo: dbinfo,
			state:  STATE_STOP,
		}
		ora.Start()
		oracles[name] = ora
		return ora
	} else {
		fmt.Printf("NewOracle Error: %s has been exist\n", name)
	}
	return nil
}

func GetOracleByName(name string) *Oracle {
	oraclemtx.Lock()
	defer oraclemtx.Unlock()
	if ora, ok := oracles[name]; ok {
		return ora
	}
	return nil
}

type OraclePool struct {
	Name      string
	DBInfo    string
	size      int
	instances []*Oracle
}

func (pool *OraclePool) GetOracle(idx int) *Oracle {
	if pool.size == 0 {
		return nil
	}
	return pool.instances[idx%pool.size]
}

func (pool *OraclePool) DBAction(cb func(*sql.DB), args ...interface{}) *Oracle {
	if pool.size == 0 {
		cb(nil)
		return nil
	}
	idx := 0
	if len(args) == 1 {
		if i, ok := args[0].(int); ok {
			idx = i % pool.size
		} else {
			idx = pool.idx % pool.size
			pool.idx++
		}
	}
	if idx < 0 {
		idx += pool.size
	}
	return pool.instances[idx%pool.size].DBAction(cb)
}

func (pool *OraclePool) GetDB(idx int) *sql.DB {
	if pool.size == 0 {
		return nil
	}
	return pool.instances[idx%pool.size].DB
}

func (pool *OraclePool) Stop() {
	for _, v := range pool.instances {
		v.Stop()
	}
}

func (pool *OraclePool) NewOraclePool(name string, dbinfo string, size int) *OraclePool {
	if name == "" {
		fmt.Printf("NewOraclePool Error: name is null\n")
		return nil
	}
	if dbinfo == "" {
		fmt.Printf("NewOraclePool Error: dbinfo is null\n")
		return nil
	}
	if pool.size == 0 {
		fmt.Printf("NewOraclePool Error: size is 0\n")
		return nil
	}
	poolmtx.Lock()
	defer poolmtx.Unlock()
	pool, ok := pools[name]
	if !ok {
		pool = &OraclePool{
			Name:   name,
			DBInfo: dbinfo,
			size:   size,
		}
		for i := 0; i < size; i++ {
			ora := NewOracle(fmt.Sprintf("%s_%d", name, i), dbinfo)
			if ora != nil {
				pool.instances = append(pool.instances, ora)
			}
		}
		pools[name] = pool
		return pool
	} else {
		fmt.Printf("NewOraclePool Error: %s has been exist\n", name)
	}
	return nil
}

func GetOraclePoolByName(name string) *OraclePool {
	poolmtx.Lock()
	defer poolmtx.Unlock()
	if pool, ok := pools[name]; ok {
		return pool
	}
	return nil
}
