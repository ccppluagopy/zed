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
	MONGO_NUM  = 20
	ACTION_NUM = 1000
)

type Student struct {
	Name string
	Age  int
}

func MongoInsert(idx int, pool *zed.MongoMgrPool, ch chan int) {
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
	zed.Init("./", "./log", true, true)
	//zed.MakeNewLogDir("./")
	const (
		TagNull = iota
		Tag1
		Tag2
		Tag3
		TagMax
	)
	var LogTags = map[int]string{
		TagNull: "zed",
		Tag1:    "Tag1",
		Tag2:    "Tag2",
		Tag3:    "Tag3",
	}
	var LogConf = map[string]int{
		//"Info":   zed.LogCmd,
		"Info":   zed.LogFile,
		"Warn":   zed.LogFile,
		"Error":  zed.LogFile,
		"Action": zed.LogFile,
	}

	zed.StartLogger(LogConf, true, TagMax, LogTags, 3, 3, 3, 3)
	for i := 0; i < 5; i++ {
		zed.LogError(Tag1, i, "log test %d", i)
	}

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
	fmt.Println(fmt.Sprintf("%d insert time used: ", COR_NUM*ACTION_NUM), time.Since(t1))
	pool.Stop()
}
