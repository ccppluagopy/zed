package observer

import (
	"encoding/json"
	"github.com/ccppluagopy/zed"
	//"sync"
	"time"
)

type ObserverClient struct {
	//sync.Mutex
	OBDelaget
	eventMgr *zed.EventMgr
	Client   *zed.TcpClient
	ticker   *time.Ticker
	running  bool
	onStop   func()
}

func (obclient *ObserverClient) heartbeat() {
	obclient.Lock()
	defer obclient.Unlock()

	if obclient.running {
		//zed.ZLog("heartbeat")
		obclient.Client.SendMsgAsync(NewNetMsg(&OBMsg{
			OP: HEARTBEAT_REQ,
		}))

	}
}

func (obclient *ObserverClient) startHeartbeat() {
	//zed.ZLog("ObserverClient start heartbeat")
	for {
		select {
		case _, ok := <-obclient.ticker.C:
			if ok {
				obclient.heartbeat()
			} else {
				break
			}
		}
	}
}

func (obclient *ObserverClient) SetCloseCB(cb func()) {
	obclient.onStop = cb
}

func (obclient *ObserverClient) Stop() {
	obclient.Lock()
	defer obclient.Unlock()
	if obclient.running {
		obclient.running = false
		obclient.ticker.Stop()
		obclient.Client.Stop()
		if obclient.onStop != nil {
			zed.NewCoroutine(func() {
				defer zed.PanicHandle(true)
				obclient.onStop()
			})
		}
	}
}

func (obclient *ObserverClient) Regist(event string, data []byte) bool {
	obclient.Lock()
	defer obclient.Unlock()

	if obclient.running {
		req := &OBMsg{
			OP:    REGIST_REQ,
			Event: event,
			Data:  data,
		}

		//req.Data = append(req.Data, data...)
		zed.ZLog("===== aaaaaaaaaa   obclient SendMsg:\n\top: %d\n\tevent: %s\n\tdata: %s", req.OP, req.Event, string(req.Data))
		obclient.Client.SendMsgAsync(NewNetMsg(req))

		return true
	}

	return false
}

func (obclient *ObserverClient) Unregist(event string, data []byte) bool {
	obclient.Lock()
	defer obclient.Unlock()
	if obclient.running {
		req := &OBMsg{
			OP:    UNREGIST_REQ,
			Event: event,
		}

		req.Data = append(req.Data, data...)
		//zed.ZLog("obclient SendMsg:\n\top: %d\n\tevent: %s\n\tdata: %s", req.OP, req.Event, string(req.Data))
		obclient.Client.SendMsgAsync(NewNetMsg(req))

		return true
	}

	return false
}

func (obclient *ObserverClient) Publish(event string, data []byte) bool {
	zed.ZLog("ObserverClient Publish, Event: %s Data: %v", event, data)

	obclient.Lock()
	defer obclient.Unlock()

	if obclient.running {
		obclient.Client.SendMsgAsync(NewNetMsg(&OBMsg{
			OP:    PUBLISH_REQ,
			Event: event,
			Data:  data,
		}))
		return true
	}

	return false
}

func (obclient *ObserverClient) Publish2(event string, data []byte) bool {
	//zed.ZLog("ObserverClient Publish, Event: %s Data: %v", event, data)

	obclient.Lock()
	defer obclient.Unlock()

	if obclient.running {
		obclient.Client.SendMsgAsync(NewNetMsg(&OBMsg{
			OP:    PUBLISH2_REQ,
			Event: event,
			Data:  data,
		}))

		return true
	}

	return false
}

//HandleMsg ...
func (obclient *ObserverClient) HandleMsg(msg *zed.NetMsg) {
	zed.Printf("ObserverClient HandleMsg, Data: %s\n", string(msg.Data))
	obclient.Lock()
	defer obclient.Unlock()

	obmsg := OBMsg{}
	err := json.Unmarshal(msg.Data, &obmsg)
	if err != nil {
		zed.Printf("ObserverClient HandleMsg Error json Unmarshal Error: %v\n", err)
		return
	}

	//opstr, ok := opname[obmsg.OP]
	_, ok := opname[obmsg.OP]
	if !ok {
		zed.Printf("ObserverClient HandleMsg Error, Invalid OP: %d\n", obmsg.OP)
		return
	}

	switch obmsg.OP {
	case OB_RSP_NONE:
		//zed.Printf("ObserverClient HandleMsg Error: %s\n", string(obmsg.Data))
		break
	case HEARTBEAT_RSP:
		//zed.Printf("ObserverClient HandleMsg: %s\n", opstr)
		break
	case REGIST_RSP:
		//zed.Printf("ObserverClient HandleMsg: %s\n", opstr)
		break
	case UNREGIST_RSP:
		//zed.Printf("ObserverClient HandleMsg: %s\n", opstr)
		break
	case PUBLISH_RSP:
		//zed.Printf("ObserverClient HandleMsg: %s\n", opstr)
		break
	case PUBLISH_NOTIFY:
		zed.Printf("ObserverClient HandleMsg PUBLISH_NOTIFY 1111111: Event: %s, Data: %s\n", obmsg.Event, string(obmsg.Data))
		obclient.eventMgr.Dispatch(obmsg.Event, obmsg.Data)
		//zed.Printf("22222  ObserverClient HandleMsg Publish: Event: %s, Data: %s\n", obmsg.Event, string(obmsg.Data))
		break
	default:
		zed.Printf("ObserverClient HandleMsg Error: No Handler\n", obmsg.OP, obclient.Client.Info())
		break
	}
}

func (obclient *ObserverClient) NewListener(tag interface{}, event interface{}, handler zed.EventHandler) bool {
	if obclient.running {
		return obclient.eventMgr.NewListener(tag, event, handler)
	}
	return false
}

func (obclient *ObserverClient) DeleteListenerInCall(tag interface{}) {
	obclient.eventMgr.DeleteListenerInCall(tag)
}

func (obclient *ObserverClient) DeleteListener(tag interface{}) {
	obclient.eventMgr.DeleteListener(tag)
}

func NewOBClient(addr string, ename string, heartbeat time.Duration) *ObserverClient {
	obclient := &ObserverClient{
		ticker:  time.NewTicker(heartbeat),
		running: true,
	}
	obclient.Client = zed.NewTcpClient(obclient, addr, 0)

	if obclient.Client != nil {
		obclient.eventMgr = zed.NewEventMgr(ename)
		if obclient.eventMgr == nil {
			zed.Println("------------------ obclient.Client:", obclient.Client)
		}
		zed.NewCoroutine(func() {
			obclient.startHeartbeat()
		})
		return obclient
	} else {
		zed.Println("===================== obclient.Client:", obclient.Client)
	}

	return nil
}
