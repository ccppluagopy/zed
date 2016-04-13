package zed

import (
	//"fmt"
	"sync"
)

type Queue struct {
	sync.Mutex
	data  []interface{}
	using bool
}

func (q *Queue) Push(item interface{}) bool {
	q.Lock()
	defer q.Unlock()

	if q.using {
		q.data = append(q.data, item)
		return true
	}

	return false
}

func (q *Queue) Pop() interface{} {
	q.Lock()
	defer q.Unlock()

	if q.using && len(q.data) > 0 {
		item := q.data[0]
		q.data = append(q.data[:0], q.data[1:]...)

		return item
	}

	return nil
}

func (q *Queue) Close() {
	q.Lock()
	defer q.Unlock()
	q.using = false
}

func NewMsgQueue() *Queue {
	return &Queue{
		using: true,
	}
}
