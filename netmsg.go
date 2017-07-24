package zed

import (
	"encoding/binary"
	"sync/atomic"
)

type NetMsgDef interface {
	Encrypt()bool
	Decrypt()bool
	GetCmd() CmdType
	SetCmd(CmdType)
	GetClient() *TcpClient
	GetMsgLen() int
	GetHeadLen() int
	GetDataLen() int
	GetData() []byte
	SetData([]byte)
	GetSendBuf() []byte
}

type NetMsg struct {
	//Cmd    CmdType
	//Len    int
	encrypted int32
	Client *TcpClient
	buf    []byte
	
}

func (msg *NetMsg) Clone() *NetMsg {
	return &NetMsg{
		Client: msg.Client,
		//Cmd:    msg.Cmd,
		buf: msg.buf,
	}
}

func (msg *NetMsg) DeepClone() *NetMsg {
	return &NetMsg{
		Client: msg.Client,
		//Cmd:    msg.Cmd,
		buf: append([]byte{}, msg.buf...),
	}
}

func (msg *NetMsg) Encrypt()bool {
	if atomic.CompareAndSwapInt32(&(msg.encrypted), 0, 1) {

	}
	return true
}

func (msg *NetMsg) Decrypt()bool {
	if atomic.CompareAndSwapInt32(&(msg.encrypted), 1, 0) {
		
	}
	return true
}

func (msg *NetMsg) GetCmd() CmdType {
	return CmdType(binary.BigEndian.Uint32(msg.buf[4:]))
}

func (msg *NetMsg) SetCmd(cmd CmdType) {
	binary.BigEndian.PutUint32(msg.buf[4:], uint32(cmd))
}

func (msg *NetMsg) GetClient() *TcpClient {
	return msg.Client
}

func (msg *NetMsg) GetData() []byte {
	return msg.buf[PACK_HEAD_LEN:]
}

func (msg *NetMsg) SetData(data []byte) {
	needLen := len(data) - len(msg.buf) + PACK_HEAD_LEN
	if needLen > 0 {
		msg.buf = append(msg.buf, make([]byte, needLen)...)
	}else if needLen < 0{
		msg.buf = msg.buf[len(data) + PACK_HEAD_LEN:]
	}
	copy(msg.buf[PACK_HEAD_LEN:], data)
}

func (msg *NetMsg) GetSendBuf() []byte {
	return msg.buf
}

func (msg *NetMsg) GetMsgLen() int {
	return len(msg.buf)
}

func (msg *NetMsg) GetHeadLen() int {
	return PACK_HEAD_LEN
}

func (msg *NetMsg) GetDataLen() int {
	return len(msg.buf) - PACK_HEAD_LEN
}

func NewNetMsg(cmd CmdType, data []byte) *NetMsg {
	msg := NetMsg{
		buf: append(make([]byte, PACK_HEAD_LEN), data...),
	}
	binary.BigEndian.PutUint32(msg.buf, uint32(len(msg.buf)-PACK_HEAD_LEN))
	binary.BigEndian.PutUint32(msg.buf[4:], uint32(cmd))
	return &msg
}
