package objmgr

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

func (mgr *ObjMgr) Get(k ObjKey) interface{} {
	container := mgr.containers[k.HashIdx()%mgr.size]
	container.RLock()
	defer container.RUnlock()

	if v, ok := container.keyvalues[k]; ok {
		return v
	}
	return nil
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