package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	port := flag.Int("port", 8800, "http port")
	flag.Parse()
	if err := Start([]string{}); err != nil {
		fmt.Println(err)
	}

	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/del", delHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/join", joinHandler)
	http.HandleFunc("/info", infoHandler)
	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}

}
