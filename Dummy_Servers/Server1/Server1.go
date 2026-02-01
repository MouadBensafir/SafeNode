package main

import (
	"net/http"
	"time"
	"fmt"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Request received !")
	time.Sleep(10 * time.Second)
	fmt.Fprintf(w, "Finished Connection !")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8081", nil)
}