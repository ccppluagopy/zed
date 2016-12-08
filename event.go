package zed

import (
	//"fmt"
	"sync"
)

var (
	defaultInstance *EventMgr = nil
	instanceMap     map[interface{}]*EventMgr
	instanceMutex   = sync.Mutex{}
)

func eventHandler(handler EventHandler, event interface{}, args []interface{}) {
	defer PanicHandle(true)
	handler(event, args)
}

func (eventMgr *EventMgr) NewListener(tag interface{}, event interface{}, handler EventHandler) bool {
	eventMgr.Lock()
	defer eventMgr.Unlock()

	if _, ok := eventMgr.listenerMap[tag]; ok {
		ZLog("NewListener Error: listener %v exist!", tag)
		return false
	}

	eventMgr.listenerMap[tag] = event
	if eventMgr.listeners[event] == nil {
		eventMgr.listeners[event] = make(map[interface{}]EventHandler)
	}
	eventMgr.listeners[event][tag] = handler

	return true
}

func (eventMgr *EventMgr) DeleteListenerInCall(tag interface{}) {
	if event, ok := eventMgr.listenerMap[tag]; ok {
		delete(eventMgr.listenerMap, tag)
		delete(eventMgr.listeners[event], tag)
	}
}

func (eventMgr *EventMgr) DeleteListener(tag interface{}) {
	eventMgr.Lock()
	defer eventMgr.Unlock()

	if event, ok := eventMgr.listenerMap[tag]; ok {
		delete(eventMgr.listenerMap, tag)
		delete(eventMgr.listeners[event], tag)
	}
}

func (eventMgr *EventMgr) dispatch(event interface{}, args ...interface{}) bool {
	flag := false
	Println("xxxxxxxxxx 000 event: ", event)
	if listeners, ok := eventMgr.listeners[event]; ok {
		for _, listener := range listeners {
			Println("xxxxxxxxxx event: ", event)
			eventHandler(listener, event, args)
			flag = true
		}
	}

	return flag
}

func (eventMgr *EventMgr) Dispatch(event interface{}, args ...interface{}) {
	eventMgr.Lock()
	defer eventMgr.Unlock()
	Println("xxxxxxxxxx 000 event: ", event)
	if listeners, ok := eventMgr.listeners[event]; ok {
		for _, listener := range listeners {
			Println("xxxxxxxxxx 111 event: ", event)
			eventHandler(listener, event, args)
		}
	}
}

func GetInstance() *EventMgr {
	if defaultInstance == nil {
		defaultInstance = &EventMgr{
			listenerMap: make(map[interface{}]interface{}),
			listeners:   make(map[interface{}]map[interface{}]EventHandler),
			//mutex:       sync.Mutex{},
			valid: true,
		}
	}
	return defaultInstance
}

func NewEventMgr(tag interface{}) *EventMgr {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()

	if _, ok := instanceMap[tag]; ok {
		ZLog("NewEventMgr Error: EventMgr %v exist!", tag)
		return nil
	}

	eventMgr := &EventMgr{
		listenerMap: make(map[interface{}]interface{}),
		listeners:   make(map[interface{}]map[interface{}]EventHandler),
		//mutex:       sync.Mutex{},
		valid: true,
	}

	if instanceMap == nil {
		instanceMap = make(map[interface{}]*EventMgr)
	}
	instanceMap[tag] = eventMgr

	return eventMgr
}

func DeleteEventMgr(tag interface{}) {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()

	if eventMgr, ok := instanceMap[tag]; ok {
		eventMgr.Lock()
		defer eventMgr.Unlock()

		for k, e := range eventMgr.listenerMap {
			if emap, ok := eventMgr.listeners[e]; ok {
				for kk, _ := range emap {
					delete(emap, kk)
				}
			}
			delete(eventMgr.listeners, k)
		}
		delete(instanceMap, tag)
	}
}

func GetEventMgrByTag(tag interface{}) (*EventMgr, bool) {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()

	if eventMgr, ok := instanceMap[tag]; ok {
		return eventMgr, true
	}
	return nil, false
}
