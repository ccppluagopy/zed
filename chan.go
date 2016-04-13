package zed

import (
	//"fmt"
	"sync"
)

type Chan struct {
	sync.Mutex
	data  chan interface{}
	using bool
}

func (ch *Chan) Push(item interface{}) bool {
	ch.Lock()
	if ch.using {
		ch.Unlock()
		ch.data <- item
		return true
	} else {
		ch.Unlock()
	}

	return false
}

func (ch *Chan) Pop() interface{} {
	ch.Lock()
	if ch.using {
		ch.Unlock()
		item, ok := <-ch.data

		if ok {
			return item
		}
	} else {
		ch.Unlock()
	}

	return nil
}

func (ch *Chan) Close() {
	ch.Lock()
	defer ch.Unlock()
	ch.using = false
}

func NewMsgChan(bufSize int) *Chan {
	return &Chan{
		using: true,
		data:  make(chan interface{}, bufSize),
	}
}
