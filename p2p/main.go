package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	username string
	password string
)

func main() {
	port := flag.Int("port", 8800, "http port")
	dbpath := flag.String("dbpath", "./db", "db path")

	flag.Parse()
	var err error
	db, err = leveldb.OpenFile(*dbpath, nil)

	if err != nil {
		fmt.Println("db error:", err)
		os.Exit(0)
	}

	fsh := http.FileServer(http.Dir("./web/asset"))
	http.Handle("/asset/", http.StripPrefix("/asset/", fsh))

	http.HandleFunc("/put", putHandler)
	http.HandleFunc("/del", delHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/join", joinHandler)
	http.HandleFunc("/info", infoHandler)
	http.HandleFunc("/kv", kv)
	http.HandleFunc("/", dashboard)
	// http.HandleFunc("/", basicAuth(dashboard))
	http.HandleFunc("/start", start)
	http.HandleFunc("/stop", stop)

	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}

}
