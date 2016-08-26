package zed

import (
	"time"
)

type CmdType uint32

type ClientIDType string

type NewConnCB func(client *TcpClient)

const (
	NullId = "Null"
)

/*type ClientIDType uint32

const (
	NullId = 0
)*/

const (
	PACK_HEAD_LEN = 8

	RECV_BUF_LEN     = 1024
	SEND_BUF_LEN     = 1024
	KEEP_ALIVE_TIME  = time.Second * 60
	READ_BLOCK_TIME  = time.Second * 65
	WRITE_BLOCK_TIME = time.Second * 5

/*	DIAL_INTERNAL  = time.Second
	MAX_DIAL_TIMES = 10*/
)

type NetMsg struct {
	Cmd    CmdType
	Len    uint32
	Client *TcpClient
	Data   []byte
}
