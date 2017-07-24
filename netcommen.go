package zed

import (
	"time"
)



const (
	PACK_HEAD_LEN = 8

	DEFAULT_RECV_BUF_LEN    = 1024
	DEFAULT_SEND_BUF_LEN    = 1024
	DEFAULT_MAX_PACK_LEN    = 1024 * 16
	DEFAULT_KEEP_ALIVE_TIME = time.Second * 60
	DEFAULT_RECV_BLOCK_TIME = time.Second * 65
	DEFAULT_SEND_BLOCK_TIME = time.Second * 5

/*	DIAL_INTERNAL  = time.Second
	MAX_DIAL_TIMES = 10*/
)
