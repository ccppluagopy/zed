package zed

import (
	/*"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"gopkg.in/mgo.v2"*/
	//"github.com/ccppluagopy/zed/zsync"
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
type WTimerCallBack func(timer *WTimer)

//type DBErrorHandler func()

type MsgHandler func(msg *NetMsg) bool

type ClientCloseCB func(client *TcpClient)

type EventHandler func(event interface{}, args []interface{})

type EventMgr struct {
	listenerMap map[interface{}]interface{}
	listeners   map[interface{}]map[interface{}]EventHandler
	sync.Mutex
	valid bool
}

type NetMsg struct {
	Cmd    CmdType
	Len    int
	Client *TcpClient
	Data   []byte
}

type ZClientDelegate interface {
	ShowClientData() bool
	MaxPackLen() int
	RecvBufLen() int
	SendBufLen() int
	RecvBlockTime() time.Duration
	SendBlockTime() time.Duration
	AliveTime() time.Duration

	RecvMsg(*TcpClient) *NetMsg
	SendMsg(*NetMsg) bool
	HandleMsg(*NetMsg)
}

type TcpClient struct {
	sync.Mutex
	conn   *net.TCPConn
	parent ZTcpClientDelegate
	ID     ClientIDType
	Idx    int
	Addr   string
	//Data    interface{}
	chSend          chan *AsyncMsg
	closeCB         map[interface{}]ClientCloseCB
	Valid           bool
	running         bool
	EnableReconnect bool
	onConnected     func(*TcpClient)
}

type TcpServer struct {
	sync.RWMutex
	running           bool
	showClientData    bool
	ClientNum         int
	listener          *net.TCPListener
	handlerMap        map[CmdType]MsgHandler
	clients           map[*TcpClient]*TcpClient
	msgFilter         func(*NetMsg) bool
	onNewConnCB       func(client *TcpClient)
	onStopCB          func()
	maxPackLen        int
	recvBlockTime     time.Duration
	recvBufLen        int
	sendBlockTime     time.Duration
	sendBufLen        int
	aliveTime         time.Duration
	delegate          ZTcpClientDelegate
	dataInSupervisor  func(*NetMsg)
	dataOutSupervisor func(*NetMsg)

	//clientIdMap map[*TcpClient]ClientIDType
	//idClientMap map[ClientIDType]*TcpClient
}
