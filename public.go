package zed

import (
	/*"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"gopkg.in/mgo.v2"*/
	"net"
	"sync"
	"time"
)

const (
	LogCmd  = 1
	LogFile = 2

	//NullID = 0
	NullID = "NULL"

	TAG_NULL = "--"
)

type CmdType uint32

//type ClientIDType uint32
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
	chTimer chan *WTimer
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
	Len    int
	Client *TcpClient
	Data   []byte
}

type TcpClient struct {
	sync.RWMutex
	conn   *net.TCPConn
	parent *TcpServer
	ID     ClientIDType
	Idx    int
	Addr   string
	//Data    interface{}
	chSend  chan *AsyncMsg
	closeCB map[interface{}]ClientCloseCB
	Valid   bool
	running bool
}

type TcpServer struct {
	sync.RWMutex
	running           bool
	showClientData    bool
	ClientNum         int
	listener          *net.TCPListener
	handlerMap        map[CmdType]MsgHandler
	clients           map[int]*TcpClient
	msgFilter         func(*NetMsg) bool
	onNewConnCB       func(client *TcpClient)
	onConnCloseCB     func(client *TcpClient)
	maxPackLen        int
	recvBlockTime     time.Duration
	recvBufLen        int
	sendBlockTime     time.Duration
	sendBufLen        int
	aliveTime         time.Duration
	delegate          ZServerDelegate
	dataInSupervisor  func(*NetMsg)
	dataOutSupervisor func(*NetMsg)

	//clientIdMap map[*TcpClient]ClientIDType
	//idClientMap map[ClientIDType]*TcpClient
}
