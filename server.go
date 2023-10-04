package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// connection, sdp
var clients = make(map[*websocket.Conn]bool)

func main() {
	http.HandleFunc("/", http.FileServer(http.Dir("./client/dist/")).ServeHTTP)
	http.HandleFunc("/ws", handleRequest)
	log.Println("http server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Host, " Joined")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// close connection when function returns
	defer conn.Close()
	// add connection to clients
	clients[conn] = true
	// check if any client is connected
	if len(clients) > 1 {
		if err := conn.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"knockKnock\"}")); err != nil {
			log.Println(err)
		}
	}

	// read message from client
	for {
		// read message from client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(clients, conn)
			break
		}
		// check message for sdp or ice
		result := string(msg)
		// log.Println(result)
		subject := strings.Split(result, "~")[0]
		if subject == "offer" {
			log.Println("sdp from", r.Host)
		}
		// content := strings.Split(result, "~")[1]
		// save connection
		clients[conn] = true
		// send message to all clients
		for client := range clients {
			if client == conn {
				continue
			}
			if err := client.WriteMessage(websocket.TextMessage, []byte(result)); err != nil {
				log.Println(err)
			}
		}

	}
}
