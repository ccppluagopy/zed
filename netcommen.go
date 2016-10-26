package zed

import (
	"time"
)

/*type ClientIDType uint32

const (
	NullID = 0
)*/

const (
	PACK_HEAD_LEN = 8

	RECV_BUF_LEN    = 1024
	SEND_BUF_LEN    = 1024
	KEEP_ALIVE_TIME = time.Second * 60
	RECV_BLOCK_TIME = time.Second * 65
	SEND_BLOCK_TIME = time.Second * 5

/*	DIAL_INTERNAL  = time.Second
	MAX_DIAL_TIMES = 10*/
)
