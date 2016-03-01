package zed

type int CmdType
type ClientIDType string

const (
	NullId = "Null"
)

var (
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
	BufLen uint32
	Buf    []byte
}
