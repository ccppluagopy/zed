package zed

import (
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"gopkg.in/mgo.v2"
	"net"
	"sync"
	"time"
)

const (
	LogCmd = iota
	LogFile

	NullId = "Null"

	TAG_NULL = ""
)

type CmdType uint32

type ClientIDType string

type NewConnCB func(client *TcpClient)

type ClosureCB func()

type TimerCallBack func()

//type DBErrorHandler func()

type MsgHandler func(msg *NetMsg) bool

type ClientCloseCB func(client *TcpClient)

type TimerWheel struct {
	running bool
	//chTicker  chan time.Time
	chTimer chan *wtimer
	//chStop    chan byte
	ticker    *time.Ticker
	currWheel int64
	wheels    []wheel
}

type EventHandler func(event interface{}, args []interface{})

type EventMgr struct {
	listenerMap map[interface{}]interface{}
	listeners   map[interface{}]map[interface{}]EventHandler
	mutex       *sync.Mutex
	valid       bool
}

type NetMsg struct {
	Cmd    CmdType
	Len    uint32
	Client *TcpClient
	Data   []byte
}

type TcpClient struct {
	sync.RWMutex
	conn    *net.TCPConn
	parent  *TcpServer
	Id      ClientIDType
	Idx     int
	Addr    string
	closeCB map[interface{}]ClientCloseCB
	chSend  chan *NetMsg
	running bool
}

type TcpServer struct {
	sync.RWMutex
	running      bool
	ClientNum    int
	listener     *net.TCPListener
	newConnCBMap map[string]func(client *TcpClient)
	handlerMap   map[CmdType]MsgHandler
	clients      map[int]*TcpClient
	clientIdMap  map[*TcpClient]ClientIDType
	idClientMap  map[ClientIDType]*TcpClient
}

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
	ticker     *time.Ticker
	running    bool
	restarting bool
}

type MongoMgrPool struct {
	mgrs []*MongoMgr
}
