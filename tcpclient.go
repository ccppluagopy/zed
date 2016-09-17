package zed

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	//"sync"
	"time"
)

type AsyncMsg struct {
	msg *NetMsg
	cb  func()
}

func (client *TcpClient) Info() string {
	return fmt.Sprintf("Client(ID: %v-Addr: %s)", client.ID, client.Addr)
}

func (client *TcpClient) AddCloseCB(key interface{}, cb ClientCloseCB) {
	client.Lock()
	defer client.Unlock()
	if client.running {
		client.closeCB[key] = cb
	}
}

func (client *TcpClient) RemoveCloseCB(key interface{}) {
	client.Lock()
	defer client.Unlock()
	if client.running {
		delete(client.closeCB, key)
	}
}

func (client *TcpClient) Stop() {
	//NewCoroutine(func() {
	client.Lock()
	defer client.Unlock()

	if client.running {
		client.parent.onClientStop(client)

		client.running = false

		client.conn.Close()

		close(client.chSend)

		for _, cb := range client.closeCB {
			cb(client)
		}

		for key, _ := range client.closeCB {
			delete(client.closeCB, key)
		}

		if showClientData {
			ZLog("[Stop] %s", client.Info())
		}
	}
	//})
}

func (client *TcpClient) writer() {
	/*var (
		writeLen = 0
		buf      []byte
		err      error
		msg      *NetMsg = nil
		ok       bool    = false
	)*/

	for {
		if asyncMsg, ok := <-client.chSend; ok {
			/*if err = (*client.conn).SetWriteDeadline(time.Now().Add(WRITE_BLOCK_TIME)); err != nil {
				LogError(LOG_IDX, client.Idx, "%s SetWriteDeadline Err: %v.", client.Info(), err)
				goto Exit
			}

			buf = make([]byte, PACK_HEAD_LEN+len(msg.Data))
			binary.LittleEndian.PutUint32(buf, uint32(len(msg.Data)))
			binary.LittleEndian.PutUint32(buf[4:8], uint32(msg.Cmd))
			copy(buf[PACK_HEAD_LEN:], msg.Data)

			writeLen, err = client.conn.Write(buf)

			LogInfo(LOG_IDX, client.Idx, "Send Msg %s Cmd: %d, Len: %d, Data: %s", client.Info(), msg.Cmd, msg.Len, string(msg.Data))

			if err != nil || writeLen != len(buf) {
				goto Exit
			}*/
			client.SendMsg(asyncMsg.msg)
			if asyncMsg.cb != nil {
				//if cb, ok := (asyncMsg.cb).(func()); ok {
				func() {
					defer func() {
						recover()
					}()
					asyncMsg.cb()
				}()
				//}
			}
		} else {
			break
		}
	}
	/*
	   Exit:
	   	client.Stop()
	   	LogInfo(LOG_IDX, client.Idx, "writer Exit %s", client.Info())*/
}

func (client *TcpClient) SendMsg(msg *NetMsg) {
	client.Lock()
	defer client.Unlock()

	var (
		writeLen = 0
		buf      []byte
		err      error
	)

	if err := (*client.conn).SetWriteDeadline(time.Now().Add(WRITE_BLOCK_TIME)); err != nil {
		LogError(LOG_IDX, client.Idx, "%s SetWriteDeadline Err: %v.", client.Info(), err)
		goto Exit
	}

	buf = make([]byte, PACK_HEAD_LEN+msg.Len)
	binary.LittleEndian.PutUint32(buf, uint32(msg.Len))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(msg.Cmd))
	if msg.Len > 0 {
		copy(buf[PACK_HEAD_LEN:], msg.Data)
	}

	writeLen, err = client.conn.Write(buf)

	if showClientData {
		ZLog("[Send] %s Cmd: %d, Len: %d, Data: %s", client.Info(), msg.Cmd, msg.Len, string(msg.Data))
	}
	//LogInfo(LOG_IDX, client.Idx, "%s Send Msg Cmd: %d, Len: %d, Data: %s", client.Info(), msg.Cmd, msg.Len, string(msg.Data))

	if err == nil && writeLen == len(buf) {
		return
	}

Exit:
	client.Stop()
}

func (client *TcpClient) SendMsgAsync(msg *NetMsg, argv ...interface{}) {
	client.RLock()
	defer client.RUnlock()

	if client.running {
		asyncmsg := &AsyncMsg{
			msg: msg,
			cb:  nil,
		}
		if len(argv) > 0 {
			if cb, ok := (argv[0]).(func()); ok {
				asyncmsg.cb = cb
			}
		}
		client.chSend <- asyncmsg
	}
}

func (client *TcpClient) reader() {
	var (
		head    = make([]byte, PACK_HEAD_LEN)
		readLen = 0
		err     error
	)

	for {
		if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
			if showClientData {
				ZLog("%s SetReadDeadline Err: %v.", client.Info(), err)
			}
			goto Exit
		}

		readLen, err = io.ReadFull(client.conn, head)
		if err != nil || readLen < PACK_HEAD_LEN {
			/*if showClientData {
				ZLog("%s Read Head Err: %v.", client.Info(), err)
			}*/
			goto Exit
		}

		if err = (*client.conn).SetReadDeadline(time.Now().Add(READ_BLOCK_TIME)); err != nil {
			if showClientData {
				ZLog("%s SetReadDeadline Err: %v.", client.Info(), err)
			}
			goto Exit
		}

		var msg = &NetMsg{
			Cmd:    CmdType(binary.LittleEndian.Uint32(head[4:8])),
			Len:    int(binary.LittleEndian.Uint32(head[0:4])),
			Client: client,
		}

		if msg.Len > 0 {
			msg.Data = make([]byte, msg.Len)
			readLen, err := io.ReadFull(client.conn, msg.Data)
			if err != nil || readLen != int(msg.Len) {
				if showClientData {
					ZLog("%s Read Body Err: %v.", client.Info(), err)
				}
				goto Exit
			}
		}

		if showClientData {
			ZLog("[Recv] %s Cmd: %d, Len: %d, Data: %s", client.Info(), msg.Cmd, msg.Len, string(msg.Data))
		}
		//LogInfo(LOG_IDX, client.Idx, "Recv Msg %s Cmd: %d, Len: %d, Data: %s", client.Info(), msg.Cmd, msg.Len, string(msg.Data))

		client.parent.HandleMsg(msg)
	}

Exit:
	client.Stop()
	//LogInfo(LOG_IDX, client.Idx, "reader Exit %s", client.Info())
}

func (client *TcpClient) start() bool {
	if err := client.conn.SetKeepAlive(true); err != nil {
		if showClientData {
			ZLog("%s SetKeepAlive Err: %v.", client.Info())
		}
		return false
	}

	if err := client.conn.SetKeepAlivePeriod(KEEP_ALIVE_TIME); err != nil {
		if showClientData {
			ZLog("%s SetKeepAlivePeriod Err: %v.", client.Info(), err)
		}
		return false
	}

	if err := (*client.conn).SetReadBuffer(RECV_BUF_LEN); err != nil {
		if showClientData {
			ZLog("%s SetReadBuffer Err: %v.", client.Info(), err)
		}
		return false
	}
	if err := (*client.conn).SetWriteBuffer(SEND_BUF_LEN); err != nil {
		if showClientData {
			ZLog("%s SetWriteBuffer Err: %v.", client.Info(), err)
		}
		return false
	}

	NewCoroutine(func() {
		client.writer()
	})

	NewCoroutine(func() {
		client.reader()
	})

	if showClientData {
		ZLog("New Client Start %s", client.Info())
	}

	return true
}

func newTcpClient(parent *TcpServer, conn *net.TCPConn) *TcpClient {
	client := &TcpClient{
		conn:    conn,
		parent:  parent,
		ID:      NullID,
		Idx:     parent.ClientNum,
		Addr:    conn.RemoteAddr().String(),
		closeCB: make(map[interface{}]ClientCloseCB),
		chSend:  make(chan *AsyncMsg, 10),
		//Data:    nil,
		running: true,
	}

	return client
}
