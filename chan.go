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

func (q *Chan) Push(item interface{}) bool {
	q.Lock()
	defer q.Unlock()

	if q.using {
		q.data <- item
		return true
	}

	return false
}

func (q *Chan) Pop() interface{} {
	q.Lock()
	defer q.Unlock()

	if q.using && len(q.data) > 0 {
		item := <-q.data

		return item
	}

	return nil
}

func (q *Chan) Close() {
	q.Lock()
	defer q.Unlock()
	q.using = false
}

func NewMsgChan() *Chan {
	return &Chan{
		using: true,
	}
}
