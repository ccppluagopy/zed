/*package main

import (
	"fmt"
	"github.com/yuin/gopher-lua"
)

func main() {
	luaState := lua.NewState()
	if err := luaState.DoFile("./tmap001.lua"); err != nil {
		fmt.Println("Dofile err: ", err)
		return
	}

	lv := luaState.Get(-1) // get the value at the top of the stack
	if tbl, ok := lv.(*lua.LTable); ok {
		// lv is LTable
		fmt.Println(luaState.ObjLen(tb))
	}
}
*/

package main

import (
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"os"
	// _ "github.com/ziutek/mymysql/thrsafe" // Thread safe engine
	"fmt"
	"time"
)

var (
	MAX_INSERT = 10000
	COR_NUM    = 80
)

func main() {

	db := mysql.New("tcp", "", "127.0.0.1:3306", "root", "", "test")

	err := db.Connect()
	if err != nil {
		panic(err)
	}

	//rows, res, err := db.Query("select * from xx where name = \"gaofeng\"")
	/*	rows, res, err := db.Query("select * from xx")
		fmt.Println("rows: ", rows)
		fmt.Println("res: ", res)
		fmt.Println("err: ", err)
		if err != nil {
			panic(err)
		}

		for i, row := range rows {
			fmt.Println(fmt.Sprintf("%d: ", i), row.Str(0), row.Int(1))
		}
	*/
	/*t1 := time.Now()
	for i := 0; i < MAX_INSERT; i++ {
		//rows, res, err := db.Query("insert xx values(\"usr%d\", %d);", i+1, i%30+20)

		//rows, res, err := db.Query("insert xx values(\"usr%d\", %d);", n*(MAX_INSERT/COR_NUM)+(i+1), i%30+20)
		//fmt.Println("insert ", i, rows, res, err)
		db.Query("insert xx values(\"usr%d\", %d);", (i + 1), i%30+20)
	}*/

	ch := make(chan int)

	dbs := make([]*mysql.Conn, COR_NUM)
	for i := 0; i < COR_NUM; i++ {
		newdb := db.Clone()
		newdb.Connect()
		dbs[i] = &newdb
	}

	t1 := time.Now()
	for i := 0; i < COR_NUM; i++ {
		go Insert(dbs[i], i, &ch)
	}

	n := 0
	for {
		_ = <-ch
		n++
		if n >= COR_NUM {
			break
		}
	}
	fmt.Println(fmt.Sprintf("%d insert time used: ", MAX_INSERT), time.Since(t1))
}

func Insert(db_ *mysql.Conn, n int, ch *chan int) {
	db := *db_
	fmt.Println("Insert: ", n)
	for i := 0; i < MAX_INSERT/COR_NUM; i++ {
		//rows, res, err := db.Query("insert xx values(\"usr%d\", %d);", i+1, i%30+20)

		rows, res, err := db.Query("insert xx values(\"usr%d\", %d);", n*(MAX_INSERT/COR_NUM)+(i+1), i%30+20)
		if err != nil {
			fmt.Println("insert ", i, rows, res, err)
		}
		//db.Query("insert xx values(\"usr%d\", %d);", n*10000+(i+1), i%30+20)
	}

	*ch <- n

}
