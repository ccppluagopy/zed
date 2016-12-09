package observer

import (
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/ccppluagopy/zed"
)

const (
	PACK_HEAD_LEN = 4

	DEFAULT_RECV_BUF_LEN    = 1024
	DEFAULT_SEND_BUF_LEN    = 1024
	DEFAULT_MAX_PACK_LEN    = 1024 * 16
	DEFAULT_KEEP_ALIVE_TIME = time.Second * 60
	DEFAULT_RECV_BLOCK_TIME = time.Second * 65
	DEFAULT_SEND_BLOCK_TIME = time.Second * 5
	DEFAULT_HEART_BEAT_TIME = time.Second * 3
	DEFAULT_ASYN_BUF_LEN    = 10

/*	DIAL_INTERNAL  = time.Second
	MAX_DIAL_TIMES = 10*/
)

const (
	OB_RSP_NONE = iota
	PUBLISH_REQ
	PUBLISH_RSP
	PUBLISH_NOTIFY
	HEARTBEAT_REQ
	HEARTBEAT_RSP
	REGIST_REQ
	REGIST_RSP
	UNREGIST_REQ
	UNREGIST_RSP
)

var (
	opname = map[int]string{
		OB_RSP_NONE:    "OB_RSP_NONE",
		PUBLISH_REQ:    "PUBLISH_REQ",
		PUBLISH_RSP:    "PUBLISH_RSP",
		PUBLISH_NOTIFY: "PUBLISH_NOTIFY",
		HEARTBEAT_REQ:  "HEARTBEAT_REQ",
		HEARTBEAT_RSP:  "HEARTBEAT_RSP",
		REGIST_REQ:     "REGIST_REQ",
		REGIST_RSP:     "REGIST_RSP",
		UNREGIST_REQ:   "UNREGIST_REQ",
		UNREGIST_RSP:   "UNREGIST_RSP",
	}
)

var (
	EventNull = ""
	EventAll  = zed.EventAll

	ErrEventFlag = "Error"

	ErrJsonUnmarshall     = "Json Unmarshall Failed"
	ErrRegistEventNull    = "Regist Event is Null"
	ErrUnregistEventNull  = "Unregist Event is Null"
	ErrUnegistNotRegisted = "Unegist Not Registed"
	ErrInvalidOP          = "InvalidOP, No Handler"
)

//OBMsg ...
type OBMsg struct {
	OP    int    `json:"OP"` //client: Regist/UnRegist/Publish   server: Rsp/Publish
	Error string `json:"Error"`
	Event string `json:"Event"`
	Data  []byte `json:"Data"`
}

//NetOBMsg send to handle function
type NetOBMsg struct {
	obmessage *OBMsg
	client    *zed.TcpClient
	Obss      *ObserverServer
}

func NewNetMsg(obm *OBMsg) *zed.NetMsg {
	data, _ := json.Marshal(obm)
	return &zed.NetMsg{
		Len:  len(data),
		Data: data,
	}
}

//------------------------------------------------------------------------------

//todo unpack
func unpack(lenbytes []byte, data []byte) *OBMsg {
	obm := new(OBMsg)
	//len := int(binary.LittleEndian.Uint32(lenbytes[0:]))
	err := json.Unmarshal(data, obm)
	if err != nil {
		zed.ZLog("unpack json unmarshal failed!!!!")
		return nil
	}

	return obm
}

//pack
func (obm *OBMsg) pack() []byte {
	bytes, err := json.Marshal(obm)
	if err != nil {
		zed.ZLog("pack json marshal failed!!!!")
		return nil
	}

	buf := make([]byte, 4+len(bytes))
	binary.LittleEndian.PutUint32(buf, uint32(len(bytes)))
	if len(bytes) > 0 {
		copy(buf[4:], bytes)
	}

	return buf
}
