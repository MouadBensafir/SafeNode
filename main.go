package main

import (
	"net/http"	
)

func handleRequests(w http.ResponseWriter, r *http.Request) {

}


func main() {
	http.HandleFunc("/backend", handleRequests)
	http.ListenAndServe(":8081", nil)
}