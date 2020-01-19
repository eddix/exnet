package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/eddix/exnet"
	"github.com/eddix/exnet/addresspicker"
)

func main() {
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:3306")
	cluster := &exnet.Cluster{
		DialTimeout:   1000 * time.Millisecond,
		ReadTimeout:   1000 * time.Millisecond,
		WriteTimeout:  1000 * time.Millisecond,
		AddressPicker: addresspicker.NewRoundRobin([]net.Addr{addr}),
	}
	mysql.RegisterDial("mydb", func(addr string) (net.Conn, error) {
		conn, err := cluster.Dial("", "")
		if err != nil {
			return nil, err
		}
		_ = exnet.TraceConn(conn, os.Stderr, nil)
		return conn, nil
	})
	db, err := sql.Open("mysql", "root@mydb(localhost:3306)/test?allowNativePasswords=true")
	if err != nil {
		log.Fatal("sql.Open: ", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("db.Ping: ", err)
	}
	rows, err := db.Query("SELECT * from key_value")
	if err != nil {
		log.Fatal("db.Query: ", err)
	}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, value); err != nil {
			log.Fatal("rows.Scan: ", err)
		}
		print(key, value)
	}
}
