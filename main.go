package main

import (
    "fmt"
    "net/http"
)

import ("github.com/gorilla/websocket")

func home(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Chat App Go fonctionne")
}

func main() {
    http.HandleFunc("/", home)

    fmt.Println("Serveur lancé sur http://localhost:8080")

    http.ListenAndServe(":8080", nil)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}