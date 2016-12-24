//package test

package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	"github.com/ccppluagopy/zed/mongo"
	"gopkg.in/mgo.v2"
	"time"
)

const (
	COR_NUM    = 100
	MONGO_NUM  = 20
	ACTION_NUM = 1000
)

type Student struct {
	Name string
	Age  int
}

func MongoInsert(idx int, pool *mongo.MongoMgrPool, ch chan int) {
	idx = idx * ACTION_NUM * 10
	for i := 0; i < ACTION_NUM; i++ {
		time.Sleep(time.Second * 3)
		pool.DBAction(idx/10/ACTION_NUM, func(collection *mgo.Collection) bool {
			err := collection.Insert(&Student{fmt.Sprintf("stu_%d", idx+i), 20 + idx/10000})
			fmt.Println("Insert err: ", err)
			if err != nil {
				panic(err)
			}

			return true
		})
	}

	ch <- 0
}

func main() {
	zed.Init("./", "./log" )


	ch := make(chan int)
	pool := mongo.NewMongoMgrPool("testmongopool", "127.0.0.1:27017", "test", "students", "usr", "passwd", MONGO_NUM)

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
	fmt.Println(fmt.Sprintf("%d insert time used: ", COR_NUM*ACTION_NUM), time.Since(t1))
	pool.Stop()
}

