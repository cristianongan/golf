package socket

import "log"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
var hubBroadcastSocket *Hub

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

// func NewHub() *Hub {
// 	return &Hub{
// 		Broadcast:  make(chan []byte),
// 		Register:   make(chan *Client),
// 		Unregister: make(chan *Client),
// 		Clients:    make(map[*Client]bool),
// 	}
// }

func InitHubSocket() {
	hubBroadcastSocket = &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func GetHubSocket() *Hub {
	return hubBroadcastSocket
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			log.Println("[SOCKET] Hub Run Register")
			log.Println("[SOCKET] len clients", len(h.Clients))
			h.Clients[client] = true
		case client := <-h.Unregister:
			log.Println("[SOCKET] Hub Run Unregister")
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.send)
			}
		case message := <-h.Broadcast:
			for client := range clients {
				err := client.WriteJSON(message)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}
