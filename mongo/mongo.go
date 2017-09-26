package mongo

import (
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	"fmt"
	//"github.com/ccppluagopy/zed"
	"sync"
	"time"
)

const (
	STATE_RUNNING = iota
	STATE_RECONNECTING
	STATE_STOP
)

const (
	DB_DIAL_TIMEOUT    = time.Second * 15
	KEEPALIVE_INTERVAL = time.Hour
)

var (
	mongos   = make(map[string]*Mongo)
	mongomtx = sync.Mutex{}

	pools   = make(map[string]*MongoPool)
	poolmtx = sync.Mutex{}

	MongoErr = &MongoError{}
)

type MongoError struct {
}

func (err *MongoError) Error() string {
	return "naivefox/mongo MongoError"
}

func Recover() {
	panic(MongoErr)
}

type Mongo struct {
	sync.Mutex
	Name       string
	Session    *mgo.Session
	Collection *mgo.Collection
	addr       string
	database   string
	collection string
	usr        string
	passwd     string
	ticker     *time.Ticker
	state      int
}

func (mongo *Mongo) Ping() bool {
	mongo.Lock()
	defer mongo.Unlock()
	return mongo.Session.Ping() == nil
}

func (mongo *Mongo) startHeartbeat() {
	if mongo.state == STATE_RUNNING {
		if mongo.ticker == nil {
			time.AfterFunc(1, func() {
				mongo.ticker = time.NewTicker(KEEPALIVE_INTERVAL)
				for {
					_, ok := <-mongo.ticker.C
					if !ok {
						break
					}
					mongo.Ping()
				}
			})
		}
	}
}

func (mongo *Mongo) Connect() {
	mongo.Lock()
	defer mongo.Unlock()
	if mongo.state != STATE_STOP {
		mongo.state = STATE_RECONNECTING
		mongo.Reset()
		session, err := mgo.DialWithTimeout(mongo.addr, DB_DIAL_TIMEOUT)
		if err != nil {
			fmt.Printf("Mongo Connect To: %s Failed, Error: %v", mongo.addr, err)
			time.AfterFunc(1, func() {
				time.Sleep(time.Second * 1)
				mongo.Connect()
			})
			return
		}

		mongo.Session = session
		mongo.Collection = session.DB(mongo.database).C(mongo.collection)

		mongo.state = STATE_RUNNING
		mongo.startHeartbeat()

		fmt.Printf("Mongo(Name: %s, db: %s collection: %s Connected()\n", mongo.addr, mongo.database, mongo.collection)
	}
}

func (mongo *Mongo) Start() {
	if mongo.state == STATE_STOP {
		mongo.state = STATE_RECONNECTING
		time.AfterFunc(1, mongo.Connect)
	}
}

func (mongo *Mongo) Reset() {
	if mongo.ticker != nil {
		mongo.ticker.Stop()
		mongo.ticker = nil
	}
	if mongo.Session != nil {
		mongo.Session.Close()
		mongo.Session = nil
	}
	if mongo.Collection != nil {
		mongo.Collection = nil
	}
}

func (mongo *Mongo) Stop() {
	mongo.Lock()
	defer mongo.Unlock()
	if mongo.state != STATE_STOP {
		mongo.state = STATE_STOP
		mongo.Reset()
	}
}

func (mongo *Mongo) DBAction(cb func(*mgo.Collection)) *Mongo {
	mongo.Lock()
	defer mongo.Unlock()
	defer func() {
		if err := recover(); err != nil {
			_, ok := err.(*MongoError)
			if ok && mongo.state == STATE_RUNNING {
				time.AfterFunc(1, func() {
					mongo.Connect()
				})
			} else {
				fmt.Printf("Mongo DBAction Error: %v\n", err)
				//log stack
			}
		}
	}()
	cb(mongo.Collection)
	return mongo
}

func NewMongo(name string, addr string, dbname string, collectionname string, usr string, passwd string) *Mongo {
	if name == "" {
		fmt.Printf("NewMongo Error: name is null\n")
		return nil
	}

	if dbname == "" {
		fmt.Printf("NewMongo Error: dbname is null\n")
		return nil
	}
	if collectionname == "" {
		fmt.Printf("NewMongo Error: collectionname is null\n")
		return nil
	}

	if addr == "" {
		fmt.Printf("NewMongo addr is null, use default addr: 127.0.0.1:27017\n")
		addr = "127.0.0.1:27017"
	}

	mongomtx.Lock()
	defer mongomtx.Unlock()
	mongo, ok := mongos[name]
	if !ok {
		mongo = &Mongo{
			Name:       name,
			addr:       addr,
			database:   dbname,
			collection: collectionname,
			usr:        usr,
			passwd:     passwd,
			//chAction: nil,
			state: STATE_STOP,
		}
		fmt.Printf("NewMongo Name: %s, Addr: %s, DBName: %s, CollectionName: %s, Usr: %s, Passwd: %s\n", name, addr, dbname, collectionname, usr, passwd)
		mongo.Start()
		mongos[name] = mongo
		return mongo
	} else {
		fmt.Printf("NewMongo Error: %s has been exist\n", name)
	}

	return nil
}

func GetMongoByName(name string) *Mongo {
	mongomtx.Lock()
	defer mongomtx.Unlock()
	if mongo, ok := mongos[name]; ok {
		return mongo
	}
	return nil
}

type MongoPool struct {
	size      int
	instances []*Mongo
}

func (pool *MongoPool) GetMongo(idx int) *Mongo {
	if pool.size == 0 {
		return nil
	}
	return pool.instances[idx%pool.size]
}

func (pool *MongoPool) DBAction(idx int, cb func(*mgo.Collection)) *Mongo {
	if pool.size == 0 {
		cb(nil)
		return nil
	}

	return pool.instances[idx%pool.size].DBAction(cb)
}

func (pool *MongoPool) GetCollection(idx int) *mgo.Collection {
	return pool.instances[idx%pool.size].Collection
}

func (pool *MongoPool) Stop() {
	for _, v := range pool.instances {
		v.Stop()
	}
}

func NewMongoPool(name string, addr string, dbname string, collectionname string, usr string, passwd string, size int) *MongoPool {
	if name == "" {
		fmt.Printf("NewMongoPool Error: name is null\n")
		return nil
	}

	if dbname == "" {
		fmt.Printf("NewMongoPool Error: dbname is null\n")
		return nil
	}
	if collectionname == "" {
		fmt.Printf("NewMongoPool Error: collectionname is null\n")
		return nil
	}

	if addr == "" {
		fmt.Printf("NewMongoPool addr is null, use default addr: 127.0.0.1:27017\n")
		addr = "127.0.0.1:27017"
	}

	poolmtx.Lock()
	defer poolmtx.Unlock()
	pool, ok := pools[name]
	if !ok {
		pool = &MongoPool{
			size:      size,
			instances: make([]*Mongo, size),
		}

		for i := 0; i < size; i++ {
			mongo := NewMongo(fmt.Sprintf("%s_%d", name, i), addr, dbname, collectionname, usr, passwd)
			if mongo != nil {
				pool.instances = append(pool.instances, mongo)
			}
		}

		pools[name] = pool

		return pool
	} else {
		fmt.Printf("NewMongoPool Error: %s has been exist!", name)
	}

	return nil
}

func GetMongoPoolByName(name string) *MongoPool {
	poolmtx.Lock()
	defer poolmtx.Unlock()

	if pool, ok := pools[name]; ok {
		return pool
	}
	return nil
}
