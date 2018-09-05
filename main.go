package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tkanos/gonfig"
)

// From: http://www.golangprograms.com/example-of-golang-crud-using-mysql-from-scratch.html

/*
DROP TABLE IF EXISTS `ipsec`;
CREATE TABLE `ipsec` (
	`ipsec_id` int(20) NOT NULL AUTO_INCREMENT,
  `date_utc` int(11),
  `host` varchar(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `up` int(11) DEFAULT NULL,
  `connecting` int(11) DEFAULT NULL,
  PRIMARY KEY (`ipsec_id`),
  UNIQUE KEY `uk_ipsec` (`date_utc`, `host`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

*/

type Configuration struct {
	User     string
	Password string
	Database string
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"

	config := Configuration{}
	err := gonfig.GetConf("./config.json", &config)

	dbUser := config.User
	dbPass := config.Password
	dbName := config.Database
	db, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

type message struct {
	DateUTC    int64  `json:"date_utc"`
	Host       string `json:"host"`
	Up         int32  `json:"up"`
	Connecting int32  `json:"connecting"`
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func insert(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	defer db.Close()
	if r.Method == "POST" {
		// Read body
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Unmarshal
		var msg message
		err = json.Unmarshal(b, &msg)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		output, err := json.Marshal(msg)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		query, err := db.Prepare("INSERT INTO ipsec(date_utc, host, up, connecting) VALUES(?,?,?,?)")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = query.Exec(msg.DateUTC, msg.Host, msg.Up, msg.Connecting)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Write(output)
	}
	// defer db.Close()
}

func main() {
	log.Println("Server started on: http://localhost:8080")
	http.HandleFunc("/", home)
	http.HandleFunc("/insert", insert)
	http.ListenAndServe(":8080", nil)
}
