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
}

func (obclient *ObserverClient) heartbeat() {
	obclient.Lock()
	defer obclient.Unlock()

	zed.ZLog("heartbeat")
	obclient.Client.SendMsgAsync(NewNetMsg(&OBMsg{
		OP: HEART_BEATREQ,
	}))
}

func (obclient *ObserverClient) startHeartbeat() {
	zed.ZLog("ObserverClient start heartbeat")
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

func (obclient *ObserverClient) Stop() {
	obclient.Lock()
	defer obclient.Unlock()
	if obclient.running {
		obclient.running = false
		obclient.ticker.Stop()
		obclient.Client.Stop()
	}
}

func (obclient *ObserverClient) Regist(event string, data []byte) {
	obclient.Lock()
	defer obclient.Unlock()

	req := &OBMsg{
		OP:    REGIST_REQ,
		Event: event,
	}

	req.Data = append(req.Data, data...)
	zed.ZLog("obclient SendMsg:\n\top: %d\n\tevent: %s\n\tdata: %s", req.OP, req.Event, string(req.Data))
	obclient.Client.SendMsgAsync(NewNetMsg(req))

	return
}

func (obclient *ObserverClient) Unregist(event string, data []byte) {
	obclient.Lock()
	defer obclient.Unlock()

	req := &OBMsg{
		OP:    UNREGIST_REQ,
		Event: event,
	}

	req.Data = append(req.Data, data...)
	zed.ZLog("obclient SendMsg:\n\top: %d\n\tevent: %s\n\tdata: %s", req.OP, req.Event, string(req.Data))
	obclient.Client.SendMsgAsync(NewNetMsg(req))

	return
}

func (obclient *ObserverClient) Publish(event string, data []byte) {
	zed.ZLog("ObserverClient Publish, Event: %s Data: %v", event, data)

	obclient.Lock()
	defer obclient.Unlock()

	obclient.Client.SendMsgAsync(NewNetMsg(&OBMsg{
		OP:    PUBLISH_REQ,
		Event: event,
		Data:  data,
	}))
}

//HandleMsg ...
func (obclient *ObserverClient) HandleMsg(msg *zed.NetMsg) {
	zed.ZLog("ObserverClient HandleMsg, Data: %v", msg.Data)
	obclient.Lock()
	defer obclient.Unlock()

	obmsg := OBMsg{}
	err := json.Unmarshal(msg.Data, &obmsg)
	if err != nil {
		zed.ZLog("ObserverClient HandleMsg Error json Unmarshal Error: %v", err)
		return
	}

	opstr, ok := opname[obmsg.OP]
	if !ok {
		zed.ZLog("ObserverClient HandleMsg Error, Invalid OP: %d", obmsg.OP)
		return
	}

	switch obmsg.OP {
	case OBRSP:
		if obmsg.Event == ErrEventFlag {
			zed.ZLog("ObserverClient HandleMsg Error: %s", string(obmsg.Data))
		} else {
			zed.ZLog("ObserverClient HandleMsg %s", opstr)
		}
		break
	case PUBLISH:
		zed.ZLog("11111  ObserverClient HandleMsg Publish: Event: %s, Data: %s", obmsg.Event, string(obmsg.Data))
		obclient.eventMgr.Dispatch(obmsg.Event, obmsg.Data)
		zed.ZLog("22222  ObserverClient HandleMsg Publish: Event: %s, Data: %s", obmsg.Event, string(obmsg.Data))
		break
	default:
		zed.ZLog("ObserverClient HandleMsg Error: No Handler", obmsg.OP, obclient.Client.Info())
		break
	}
}

func (obclient *ObserverClient) NewListener(tag interface{}, event interface{}, handler zed.EventHandler) bool {
	return obclient.eventMgr.NewListener(tag, event, handler)
}

func (obclient *ObserverClient) DeleteListenerInCall(tag interface{}) {
	obclient.eventMgr.DeleteListenerInCall(tag)
}

func NewOBClient(addr string, heartbeat time.Duration) *ObserverClient {
	obclient := &ObserverClient{
		ticker:   time.NewTicker(heartbeat),
		running:  true,
		eventMgr: zed.NewEventMgr(zed.Sprintf("obc_%s", addr)),
	}
	obclient.Client = zed.NewTcpClient(obclient, addr, 0)

	if obclient.Client != nil {
		zed.NewCoroutine(func() {
			obclient.startHeartbeat()
		})
		return obclient
	}

	return nil
}
