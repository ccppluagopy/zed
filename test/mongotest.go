package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	COR_NUM    = 100
	MONGO_NUM  = 10
	ACTION_NUM = 1000
)

type Student struct {
	Name string
	Age  int
}

func MongoInsert(idx int, pool *zed.MongoMgrPool, ch chan int) {
	idx = idx * ACTION_NUM * 10
	ok := true
	for i := 0; i < ACTION_NUM; i++ {
		pool.DBAction(idx, func(collection *mgo.Collection) {
			err := collection.Insert(&Student{fmt.Sprintf("stu_%d", idx+i), 20 + idx/10000})
			if err != nil {
				fmt.Println("Insert err: ", err)
				ok = false
			}
		})

		if !ok {
			break
		}
	}

	ch <- 0
}

func main() {
	ch := make(chan int)
	pool := zed.NewMongoMgrPool("testmongopool", "127.0.0.1:27017", "test", "students", "usr", "passwd", MONGO_NUM)

	t1 := time.Now()

	if pool != nil {
		for i := 0; i < COR_NUM; i++ {
			go MongoInsert(i, pool, ch)
		}
	}

	n := 0
	for {
		_ = <-ch
		n++
		if n >= COR_NUM {
			break
		}
	}
	fmt.Println(fmt.Sprintf("%d insert time used: ", MONGO_NUM*ACTION_NUM), time.Since(t1))
}
