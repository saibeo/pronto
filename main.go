package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

//connected clients
var clients = make(map[*websocket.Conn]bool)

//broadcast channel
var broadcast = make(chan Message)

// configure upgrade
var upgrader = websocket.Upgrader{}

// Message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Messsage string `json:"message"`
}

func main() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)

	// start listening for incoming chats
	go handleMessages()

	log.Println("http server has been started on port :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// Register new clients
	clients[ws] = true

	for {
		var msg Message

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
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
