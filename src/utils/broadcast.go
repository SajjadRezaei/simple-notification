package utils

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients = make(map[*websocket.Conn]map[string]bool)
	mu      = new(sync.Mutex)
)

type BroadcastRequest struct {
	EventType string `json:"event_type"`
	Message   []byte `json:"message"`
}

func RegisterClient(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	clients[conn] = make(map[string]bool)
	log.Println("New WebSocket Client Connected:", conn.RemoteAddr())
}

func UnRegisterClient(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	delete(clients, conn)
	conn.Close()
	log.Println("WebSocket Client disconnected:")
}

func SubscribeToEvent(conn *websocket.Conn, eventType string) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := clients[conn]; !exists {
		clients[conn] = make(map[string]bool)
	}
	// Add the subscription
	clients[conn][eventType] = true
	log.Printf("Client subscribed to event: %s", eventType)
}

func UnSubscribeFromEvent(conn *websocket.Conn, eventType string) {

	mu.Lock()
	defer mu.Unlock()

	// Check if the client exists
	if subscriptions, exists := clients[conn]; exists {
		// Remove the specific subscription
		delete(subscriptions, eventType)

		// If the client has no more subscriptions, remove the client
		if len(subscriptions) == 0 {
			delete(clients, conn)
		}
		log.Printf("Client unsubscribed from event: %s", eventType)
	}
}

func BroadcastMessage(eventType string, message []byte) bool {
	mu.Lock()
	defer mu.Unlock()

	log.Printf("Broadcasting message to clients subscribed to event: %s", eventType)
	log.Printf("Current clients: %+v", clients)

	success := false

	for client, subscription := range clients {
		if _, subscribed := subscription[eventType]; subscribed {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to send message to client: %v", err)
				UnRegisterClient(client)
			} else {
				success = true
			}
		}
	}

	return success
}
