package socket

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
var HubBroadcastSocket *Hub

type Hub struct {
	// Registered Clients.
	Clients []*Client

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
		Clients:    []*Client{},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// h.Clients[client] = true
			h.Clients = append(h.Clients, client)
		case client := <-h.Unregister:
			// if _, ok := h.Clients[client]; ok {
			// 	delete(h.Clients, client)
			// 	close(client.send)
			// }

			j := 0
			for _, c := range h.Clients {
				if c != client {
					// c.Clients[j] = c
					h.Clients[j] = c
					j++
				}
			}
			h.Clients = h.Clients[:j]
		case message := <-h.Broadcast:
			for _, client := range h.Clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					j := 0
					for _, c := range h.Clients {
						if c != client {
							// c.Clients[j] = c
							h.Clients[j] = c
							j++
						}
					}
					h.Clients = h.Clients[:j]
				}
			}
		}
	}
}
