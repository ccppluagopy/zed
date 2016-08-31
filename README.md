安装mysql，启动mysql： mysqld
安装mongo，启动mongo： mongod --dbpath D:\cache\Mongodb\data

deps:
	go get gopkg.in/mgo.v2
	go get github.com/go-sql-driver/mysql
	go get github.com/ziutek/mymysql/mysql

setup:
	go get github.com/ccppluagopy/zed