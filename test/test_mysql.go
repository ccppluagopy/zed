package main

import (
	//"fmt"
	"github.com/ccppluagopy/zed"
	zmysql "github.com/ccppluagopy/zed/mysql"
	"github.com/ziutek/mymysql/mysql"
	"os"
	"time"
)

const (
	Tag1 = iota
	Tag2
	Tag3
)

func main() {
	msql := zmysql.NewMysql("test", "127.0.0.1:3306", "test", "root", "")
	for {
		time.Sleep(time.Second)
		msql.DBAction(func(conn mysql.Conn) {
			_, _, err := conn.Query("select * from company")
			zed.Println(" rows: ", err)
			if err != nil {
				panic(err)
			}

		})
	}
	zed.WaitSignal(true, func(sig os.Signal) {
		os.Exit(0)
	})
	time.Sleep(time.Hour)
}
