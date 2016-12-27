package observer

import (
	"encoding/binary"
	"io"
	"time"

	"github.com/ccppluagopy/zed"
)

type OBDelaget struct {
	zed.DefaultTCDelegate
}

func (dele *OBDelaget) RecvMsg(client *zed.TcpClient) *zed.NetMsg {
	//zed.ZLog("OBDelaget RecvMsg %s", client.GetConn().LocalAddr().String())
	var (
		head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error
		msg     *zed.NetMsg
	)

	if err = client.GetConn().SetReadDeadline(time.Now().Add(DEFAULT_RECV_BLOCK_TIME)); err != nil {
		if dele.ShowClientData() {
			zed.ZLog("OBsever RecvMsg %s SetReadDeadline Err: %v.", client.Info(), err)
		}
		goto Exit
	}

	readLen, err = io.ReadFull(client.GetConn(), head)
	if err != nil || readLen < PACK_HEAD_LEN {
		if dele.ShowClientData() {
			zed.ZLog("OBsever RecvMsg %s Read Head Err: %v %d.", client.Info(), err, readLen)
		}
		goto Exit
	}

	if err = client.GetConn().SetReadDeadline(time.Now().Add(DEFAULT_RECV_BLOCK_TIME)); err != nil {
		if dele.ShowClientData() {
			zed.ZLog("OBsever RecvMsg %s SetReadDeadline Err: %v.", client.Info(), err)
		}

		goto Exit
	}

	//zed.ZLog("OBsever recved head: %v.", head)

	msg = &zed.NetMsg{
		Cmd:    zed.CmdType(0),
		Len:    int(binary.LittleEndian.Uint32(head[0:PACK_HEAD_LEN])),
		Client: client,
	}
	if msg.Len > DEFAULT_MAX_PACK_LEN {
		if dele.ShowClientData() {
			zed.ZLog("OBsever RecvMsg Read Body Err: Body Len(%d) > MAXPACK_LEN(%d)", msg.Len, DEFAULT_MAX_PACK_LEN)
		}
		goto Exit
	}
	if msg.Len > 0 {
		msg.Data = make([]byte, msg.Len)
		readLen, err := io.ReadFull(client.GetConn(), msg.Data)
		if err != nil || readLen != int(msg.Len) {
			if dele.ShowClientData() {
				zed.ZLog("OBsever RecvMsg %s Read Body Err: %v.", client.Info(), err)
			}
			goto Exit
		}
	}

	//zed.ZLog("recvmsg from %v", client.GetConn().RemoteAddr())
	return msg

Exit:
	return nil
}

func (dele *OBDelaget) SendMsg(client *zed.TcpClient, msg *zed.NetMsg) bool {
	//zed.ZLog("enter SendMsg function ....")
	var (
		writeLen = 0
		buf      []byte
		err      error
	)

	if msg.Len > 0 && (msg.Data == nil || msg.Len != len(msg.Data)) {
		if dele.ShowClientData() {
			zed.ZLog("SendMsg Err: msg.Len(%d) != len(Data)%v", msg.Len, len(msg.Data))
		}
		goto Exit
	}

	if msg.Len > DEFAULT_MAX_PACK_LEN {
		if dele.ShowClientData() {
			zed.ZLog("SendMsg Err: Body Len(%d) > MAXPACK_LEN(%d)", msg.Len, DEFAULT_MAX_PACK_LEN)
		}
		goto Exit
	}

	if err = client.GetConn().SetWriteDeadline(time.Now().Add(DEFAULT_SEND_BLOCK_TIME)); err != nil {
		if dele.ShowClientData() {
			zed.ZLog("%s SetWriteDeadline Err: %v.", client.Info(), err)
		}
		goto Exit
	}

	buf = make([]byte, PACK_HEAD_LEN+msg.Len)
	binary.LittleEndian.PutUint32(buf, uint32(msg.Len))
	if msg.Len > 0 {
		copy(buf[PACK_HEAD_LEN:], msg.Data)
	}

	//zed.ZLog("OBDelegate %s send msg to %s.", client.GetConn().LocalAddr().String(), client.GetConn().RemoteAddr().String())
	writeLen, err = client.GetConn().Write(buf)

	if err == nil && writeLen == len(buf) {
		return true
	}

Exit:
	client.Stop()
	return false
}
