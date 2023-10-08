package websocket

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type ClusterSocket struct {
	socket *websocket.Upgrader
}

func CreateClusterSocket() *ClusterSocket {
	socket := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return &ClusterSocket{socket: &socket}
}

func (socket *ClusterSocket) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := socket.socket.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not establish WebSocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return
		}
		// Handle the WebSocket message (e.g., echo it back to the client)
		err = conn.WriteMessage(messageType, p)
		if err != nil {
			return
		}
	}
}
