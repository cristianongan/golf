package socket

import "log"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
var HubBroadcastSocket *Hub

type Hub struct {
	// Registered Clients.
	Clients map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.send)
			}
		case message := <-h.Broadcast:
			log.Println("[SOCKET] Hub Run message " + string(message))
			for client := range h.Clients {

				select {
				case client.send <- message:
					log.Println("[SOCKET] Hub Run client.Send message " + string(message))
					log.Println("[SOCKET] Hub Run client.Send " + string(<-client.send))
				default:
					log.Println("[SOCKET] Hub Run default client.Send " + string(<-client.send))
					// close(client.send)
					// delete(h.Clients, client)
				}
			}
		}
	}
}
