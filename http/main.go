package main

import (
	"encoding/json"
	"net/http"
)

func msg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := string("The code works.. Fr")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/", msg)
	port := ":3155"
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	http.ListenAndServe(port, nil)
}
