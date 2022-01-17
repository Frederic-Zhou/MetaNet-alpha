package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	port := flag.Int("port", 8800, "http port")
	dbpath := flag.String("dbpath", "./db", "db path")

	flag.Parse()
	var err error
	db, err = leveldb.OpenFile(*dbpath, nil)
	if err != nil {
		os.Exit(4)
	}

	fsh := http.FileServer(http.Dir("./web/asset"))
	http.Handle("/asset/", http.StripPrefix("/asset/", fsh))

	http.HandleFunc("/put", putHandler)
	http.HandleFunc("/directline", directlineHandler)
	http.HandleFunc("/del", delHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/sendto", sendtoHandler)
	http.HandleFunc("/join", joinHandler)
	http.HandleFunc("/info", infoHandler)
	http.HandleFunc("/kv", kv)
	http.HandleFunc("/", dashboard)
	// http.HandleFunc("/", basicAuth(dashboard))
	http.HandleFunc("/start", start)
	http.HandleFunc("/stop", stop)
	http.HandleFunc("/errorlog", errorlog)

	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}

}
