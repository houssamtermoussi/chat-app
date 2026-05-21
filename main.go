package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// ---------------------------------------------------------------------------
// Point d'entrée : démarre le serveur HTTP et la goroutine du Hub
// ---------------------------------------------------------------------------

func main() {
	hub := newHub()
	go hub.run() // goroutine unique qui gère inscriptions + broadcast

	http.HandleFunc("/", home)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})

	log.Println("Serveur lancé sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<!DOCTYPE html>
<html><body>
<h1>Chat Go</h1>
<p>Ouvre la console du navigateur ou utilise un client WebSocket sur <code>ws://localhost:8080/ws</code></p>
<script>
const ws = new WebSocket("ws://" + location.host + "/ws");
ws.onmessage = (e) => {
  const p = document.createElement("p");
  p.textContent = e.data;
  document.body.appendChild(p);
};
ws.onopen = () => ws.send("Bonjour depuis le navigateur !");
</script>
</body></html>`)
}

// ---------------------------------------------------------------------------
// WebSocket : upgrade HTTP → WS, puis délègue à readPump / writePump
// ---------------------------------------------------------------------------

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // en prod : vérifier l'origine (CORS)
	},
}

func serveWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256), // tampon : évite de bloquer le Hub
	}

	hub.register <- client

	// Chaque client a 2 goroutines :
	// - readPump  : lit le WS → envoie au broadcast
	// - writePump : lit le channel send → écrit sur le WS
	go client.writePump()
	go client.readPump()
}
