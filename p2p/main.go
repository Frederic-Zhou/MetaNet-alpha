package main

import (
	"fmt"
	"net/http"
)

func main() {

	if err := Start(); err != nil {
		fmt.Println(err)
	}

	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/del", delHandler)
	http.HandleFunc("/get", getHandler)
	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}

}
