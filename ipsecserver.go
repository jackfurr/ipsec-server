package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tkanos/gonfig"
)

// https://cloud.docker.com/swarm/jackefurr/repository/docker/jackefurr/ipsecserver/general

// sudo docker run --restart --name ipsecserver -p 80:3001 -d --rm jackefurr/ipsecserver:4

// select * from ipsec where date_utc > UNIX_TIMESTAMP(utc_timestamp())-65 order by host;

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

type database struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

type configuration struct {
	Database   database
	ListenPort int
}

var (
	config = configuration{}
)

var tmpl = template.Must(template.ParseGlob("templates/*"))

func init() {
	err := gonfig.GetConf("./config.json", &config)
	if err != nil {
		panic(err.Error())
	}
	log.Println(config)
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"

	dbUser := config.Database.User
	dbPass := config.Database.Password
	dbHost := config.Database.Host
	dbPort := config.Database.Port
	dbName := config.Database.Database
	dnConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	log.Println(dnConnectionString)
	db, err := sql.Open(dbDriver, dnConnectionString)
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

type IPSecData struct {
	DateUTC    int
	Host       string
	Up         int
	Connecting int
}

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		host := r.URL.Query().Get("host")
		startUTC := r.URL.Query().Get("start_utc")
		endUTC := r.URL.Query().Get("end_utc")

		db := dbConn()
		// defer db.Close()

		selDB, err := db.Query("SELECT date_utc, host, up, connecting FROM ipsec WHERE host=? AND date_utc >= ? AND date_utc <= ? ORDER BY date_utc", host, startUTC, endUTC)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		ipSecRow := IPSecData{}
		res := []IPSecData{}
		for selDB.Next() {
			var date_utc, up, connecting int
			var host string
			err = selDB.Scan(&date_utc, &host, &up, &connecting)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			ipSecRow.DateUTC = date_utc
			ipSecRow.Host = host
			ipSecRow.Up = up
			ipSecRow.Connecting = connecting
			res = append(res, ipSecRow)
		}
		w.Header().Set("content-type", "application/json")
		tmpl.ExecuteTemplate(w, "Get", res)
		defer db.Close()
	}
}

func main() {
	log.Println(fmt.Sprintf("Server started on: http://localhost:%v", config.ListenPort))
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	// http.HandleFunc("/", home)
	http.HandleFunc("/insert", insert)
	http.HandleFunc("/get", get)
	http.ListenAndServe(fmt.Sprintf(":%v", config.ListenPort), nil)
}
