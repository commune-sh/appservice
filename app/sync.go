package app

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"maunium.net/go/mautrix/event"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SyncClient struct {
	ID    string
	Rooms map[string]bool
	Conn  *websocket.Conn
}

var clients = make(map[string]*SyncClient)
var Broadcast = make(chan *event.Event)
var mutex sync.Mutex

func (c *App) Sync() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade to websocket: %v", err)
			return
		}
		defer conn.Close()

		clientID := r.URL.Query().Get("client_id")
		roomID := r.URL.Query().Get("room_id")
		client := &SyncClient{ID: clientID, Conn: conn, Rooms: map[string]bool{roomID: true}}

		mutex.Lock()
		clients[clientID] = client
		mutex.Unlock()

		for {
			var message map[string]string
			err := conn.ReadJSON(&message)
			if err != nil {
				log.Printf("Client disconnected: %v", err)
				mutex.Lock()
				delete(clients, clientID)
				mutex.Unlock()
				break
			}

			if newRoomID, ok := message["room_id"]; ok {
				mutex.Lock()
				client.Rooms[newRoomID] = true
				mutex.Unlock()
			}
		}
	}
}

func (c *App) HandleBroadcast() {
	for {
		event := <-Broadcast
		mutex.Lock()
		for _, client := range clients {
			if client.Rooms[event.RoomID.String()] {
				err := client.Conn.WriteJSON(event)
				if err != nil {
					log.Printf("Error broadcasting to client %s: %v", client.ID, err)
					client.Conn.Close()
					delete(clients, client.ID)
				}
			}
		}
		mutex.Unlock()
	}
}
