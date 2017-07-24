package zed

import (
	"encoding/binary"
)

type NetMsgDef interface {
	GetCmd() CmdType
	SetCmd(CmdType)
	GetClient() *TcpClient
	MsgLen() int
	HeadLen() int
	DataLen() int
	GetData() []byte
	SetData([]byte)
	GetSendBuf() []byte
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
	return msg.buf
}

func (msg *NetMsg) SetData(data []byte) {
	needLen := len(data) - len(msg.buf) + PACK_HEAD_LEN
	if needLen > 0 {
		msg.buf = append(msg.buf, make([]byte, needLen)...)
	}
	copy(msg.buf[PACK_HEAD_LEN:], data)
}

func (msg *NetMsg) GetSendBuf() []byte {
	return msg.buf
}

func (msg *NetMsg) MsgLen() int {
	return len(msg.buf)
}

func (msg *NetMsg) HeadLen() int {
	return PACK_HEAD_LEN
}

func (msg *NetMsg) DataLen() int {
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
