package main

import (
	"fmt"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "server1: request received ")
	time.Sleep(3 * time.Second)
	fmt.Fprintf(w, "| server1: finished")
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/health", health)
	http.ListenAndServe(":9001", nil)
}