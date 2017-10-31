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

	TAG_NULL = "--"
)

//type ClientIDType uint32
//type ClientIDType string

type NewConnCB func(client *TcpClient)

type ClosureCB func()

type TimerCallBack func()
type WTimerCallBack func(timer *WTimer)

//type DBErrorHandler func()

type MsgHandler func(msg INetMsg) bool

type ClientCloseCB func(client *TcpClient)

type EventHandler func(event interface{}, args []interface{})

type EventMgr struct {
	listenerMap map[interface{}]interface{}
	listeners   map[interface{}]map[interface{}]EventHandler
	sync.Mutex
	valid bool
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
	parent ITcpClientDelegate
	//ID     ClientIDType
	Idx  int
	Addr string
	//Data    interface{}
	chSend          chan *AsyncMsg
	closeCB         map[interface{}]ClientCloseCB
	Valid           bool
	running         bool
	EnableReconnect bool
	onConnected     func(*TcpClient, bool)
	UserData        interface{}
}

type TcpServer struct {
	sync.RWMutex
	running           bool
	showClientData    bool
	ClientNum         int
	listener          *net.TCPListener
	handlerMap        map[uint32]MsgHandler
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
	delegate          ITcpClientDelegate
	dataInSupervisor  func(*NetMsg)
	dataOutSupervisor func(*NetMsg)

	//clientIdMap map[*TcpClient]ClientIDType
	//idClientMap map[ClientIDType]*TcpClient
}
