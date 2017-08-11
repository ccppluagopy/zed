package zed

import (
	"sync"
)

type ObjKey interface {
	HashIdx() int
}

type objcontainer struct {
	sync.RWMutex
	keyvalues map[ObjKey]interface{}
}

type ObjMgr struct {
	containers []*objcontainer
	size       int
}

func (mgr *ObjMgr) Get(k ObjKey) (interface{},bool) {
	container := mgr.containers[k.HashIdx()%mgr.size]
	container.RLock()
	defer container.RUnlock()

	if v, ok := container.keyvalues[k]; ok {
		return v, true
	}
	return nil, false
}

func (mgr *ObjMgr) Set(k ObjKey, v interface{}) {
	container := mgr.containers[k.HashIdx()%mgr.size]
	container.Lock()
	defer container.Unlock()
	container.keyvalues[k] = v
}

func (mgr *ObjMgr) Delete(k ObjKey) {
	container := mgr.containers[k.HashIdx()%mgr.size]
	container.Lock()
	defer container.Unlock()
	delete(container.keyvalues, k)
}


func (mgr *ObjMgr) Size() int {
	size := 0
	for _, container := range mgr.containers {
		container.RLock()
		defer container.RUnlock()
		size += len(container.keyvalues)
	}
	return size
}

func (mgr *ObjMgr) ForEach(cb func(ObjKey, interface{})) {
	defer HandlePanic(true)
	for _, container := range mgr.containers {
		container.RLock()
		defer container.RUnlock()
		for k, v := range container.keyvalues {
			cb(k, v)
		}
	}
}

func (mgr *ObjMgr) ForEachAsync(cb func(ObjKey, interface{})) {
	Async(func(){
		for _, container := range mgr.containers {
			//Async(func() {
				container.RLock()
				defer container.RUnlock()
				for k, v := range container.keyvalues {
					cb(k, v)
				}
			//})
		}
	})
}

func (mgr *ObjMgr) Init(bucketNum int) *ObjMgr {
	mgr.size = bucketNum
	mgr.containers = make([]*objcontainer, bucketNum)
	
	for i := 0; i < bucketNum; i++ {
		mgr.containers[i] = &objcontainer{
			keyvalues: make(map[ObjKey]interface{}),
		}
	}

	return mgr
}

func NewObjMgr(bucketNum int) *ObjMgr {
	mgr := &ObjMgr{
		size:       bucketNum,
		containers: make([]*objcontainer, bucketNum),
	}
	for i := 0; i < bucketNum; i++ {
		mgr.containers[i] = &objcontainer{
			keyvalues: make(map[ObjKey]interface{}),
		}
	}

	return mgr
}
