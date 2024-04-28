package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "ls-2c79d0fc9b182b470722892b6096eb1e6efbd8d6.czkewaw2wt17.ap-south-1.rds.amazonaws.com"
	port     = 5432
	user     = "dbuser"
	password = "+_6dSmiz!dZj2:+&RtrSMQ6DDm.kbij5"
	dbname   = "gocrud"
)

var db *sql.DB

func Init() (*sql.DB, error) {
	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)

	// open database
	var err error
	db, err = sql.Open("postgres", psqlconn)
	if err != nil {
		return nil, err
	}

	// check db
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to PostgreSQL database!")
	return db, nil
}
