package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// ---------------------------------------------------------------------------
// Hub : cœur du chat — une seule goroutine (run) modifie la map clients
// ---------------------------------------------------------------------------
//
// Pourquoi un Hub ?
//   - La map clients n'est jamais lue/écrite depuis plusieurs goroutines
//     en même temps → pas de data race, pas de crash aléatoire.
//   - register / unregister / broadcast passent par des channels :
//     pattern classique en Go pour partager l'état en toute sécurité.

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// run tourne dans une goroutine lancée depuis main().
// Boucle infinie : attend un événement sur un des 3 channels.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connecté (%d en ligne)", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send) // signale à writePump de s'arrêter
				log.Printf("Client déconnecté (%d en ligne)", len(h.clients))
			}

		case message := <-h.broadcast:
			// Copie des clients à notifier (on modifie la map pendant l'envoi)
			for client := range h.clients {
				select {
				case client.send <- message:
					// message mis en file d'attente pour writePump
				default:
					// client trop lent ou déjà mort → on le retire
					close(client.send)
					delete(h.clients, client)
					client.conn.Close()
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Client : une connexion WebSocket + un channel pour les messages sortants
// ---------------------------------------------------------------------------

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// readPump lit les messages entrants et les envoie au broadcast.
// À la fermeture (erreur ou déconnexion), demande au Hub de retirer ce client.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read: %v", err)
			}
			break
		}
		c.hub.broadcast <- message
	}
}

// writePump envoie au client tout ce que le Hub met dans send.
// Un ping périodique garde la connexion vivante derrière les proxies.
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub a fermé send → on quitte proprement
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
