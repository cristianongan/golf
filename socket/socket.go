package socket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var Broadcast = make(chan any)               // broadcast channel
var Broadcast1 = make(chan string)           // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		var msg string
		// Read in a new message as JSON and map it to a Message object

		err := ws.ReadJSON(&msg)

		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		// Send the newly received message to the broadcast channel
		Broadcast1 <- msg
	}
}

func HandleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-Broadcast
		// Send it out to every client that is currently connected

		log.Println("HandleMessages:", clients)

		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
