package main

import (
	"log"
	"net/http"
)

func main() {
	manager := NewManager()

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.serveWS)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
