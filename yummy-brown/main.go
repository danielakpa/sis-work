package main

import (
	"fmt"
	"net/http"
)

func MyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Gracy Foodie Production")
}

func main() {
	http.HandleFunc("/", MyHandler)
	http.ListenAndServe(":8080", nil)
}
