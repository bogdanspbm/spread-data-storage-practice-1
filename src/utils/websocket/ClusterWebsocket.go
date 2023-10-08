package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
)

type ClusterSocket struct {
	socket *websocket.Upgrader
}

type SocketMessage struct {
	EType string `json:"eType"`
	EName string `json:"eName"`
}

func (message SocketMessage) toString() []byte {
	jsonString, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	return jsonString
}

func (message SocketMessage) processMessage(conn *websocket.Conn) {
	fmt.Println(string(message.toString()))

	switch message.EType {
	case "connection":
		response := SocketMessage{"connection_response", "Connection Accepted"}
		conn.WriteMessage(websocket.TextMessage, response.toString())
	}
}

func CreateClusterSocket() *ClusterSocket {
	socket := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin
		},
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

		switch messageType {
		case websocket.TextMessage:
			{
				var message SocketMessage

				err := json.Unmarshal(p, &message)
				if err == nil && message.EType != "" {
					message.processMessage(conn)
					break
				}

				var response = SocketMessage{"exception", "Unsupported Message Type"}
				err = conn.WriteMessage(websocket.TextMessage, response.toString())
			}

		}

	}
}

func (socket *ClusterSocket) ConnectToNode(port int) {
	addr := fmt.Sprintf("ws://127.0.0.1:%v/ws", port)

	u, err := url.Parse(addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	eventMessage := SocketMessage{"connect", "New Cluster Connected"}

	err = c.WriteMessage(websocket.TextMessage, eventMessage.toString())

	if err != nil {
		fmt.Println(err)
		return
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Received: %s\n", message)
		}
	}()

	c.SetCloseHandler(func(code int, text string) error {
		fmt.Printf("WebSocket closed: %v (%s)\n", code, text)
		return nil
	})
}
