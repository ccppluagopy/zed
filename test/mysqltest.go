//package test

package main

import (
	"fmt"
	"github.com/ccppluagopy/zed"
	zmysql "github.com/ccppluagopy/zed/mysql"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"sync"
	"time"
)

const (
	COR_NUM    = 10
	MYSQL_NUM  = 3
	ACTION_NUM = 1000000
)

type Student struct {
	Name string
	Age  int
}

func MysqlInsert(idx int, pool *zmysql.MysqlMgrPool, wg *sync.WaitGroup) {
	defer wg.Done()

	n := idx
	idx = idx * ACTION_NUM * 10

	for i := 0; i < ACTION_NUM/COR_NUM; i++ {
		time.Sleep(time.Second * 3)
		pool.DBAction(n, func(db *mysql.Conn) {
			_, _, err := (*db).Query("insert xx values(\"usr%d\", %d);", n*ACTION_NUM+(i+1), i%30+20)

			fmt.Println("insert err:", n, err)
			if err != nil {

				panic(err)
			}

		})

	}
}

func main() {
	zed.Init("./", "./log")


	pool := zmysql.NewMysqlMgrPool("testmysqlpool", "127.0.0.1:3306", "test", "root", "", MYSQL_NUM)
	fmt.Println("pool: ", pool)
	t1 := time.Now()

	wg := new(sync.WaitGroup)
	wg.Add(COR_NUM)
	if pool != nil {
		for i := 0; i < COR_NUM; i++ {
			go MysqlInsert(i, pool, wg)
		}
	}

	wg.Wait()
	fmt.Println(fmt.Sprintf("%d insert time used: ", COR_NUM*ACTION_NUM), time.Since(t1))
	pool.Stop()
}
