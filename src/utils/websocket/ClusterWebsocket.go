package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"spread-data-storage-practice-1/src/utils/adapters"
)

type ClusterSocket struct {
	database    *adapters.DatabaseAdapter
	socket      *websocket.Upgrader
	connections []*websocket.Conn
}

func CreateClusterSocket(database *adapters.DatabaseAdapter) *ClusterSocket {
	socket := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin
		},
	}
	return &ClusterSocket{database: database, socket: &socket, connections: make([]*websocket.Conn, 4)}
}

type SocketMessage struct {
	EType      string   `json:"eType"`
	EName      string   `json:"eName"`
	EArguments []string `json:"EArguments"`
}

func (message SocketMessage) toString() []byte {
	jsonString, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	return jsonString
}

func (socket *ClusterSocket) processMessage(message SocketMessage, conn *websocket.Conn) {
	fmt.Println(string(message.toString()))

	switch message.EType {
	case "connection":
		response := SocketMessage{"connection_response", "Connection Accepted", nil}
		conn.WriteMessage(websocket.TextMessage, response.toString())
		socket.ReplicateAllValues(conn)
	case "full_replication":
		for i := 0; i < len(message.EArguments)-1; i += 2 {
			key := message.EArguments[i]
			value := message.EArguments[i+1]

			socket.database.SetValue(key, value)
		}
	case "replicate_value":
		if len(message.EArguments) < 2 {
			response := SocketMessage{"replicate_value_response", "Bad Arguments", nil}
			conn.WriteMessage(websocket.TextMessage, response.toString())
			break
		}

		key := message.EArguments[0]
		value := message.EArguments[1]

		socket.database.SetValue(key, value)
	}
}

func (socket *ClusterSocket) ReplicateValue(key string, value string) {
	for _, conn := range socket.connections {
		if conn == nil {
			continue
		}
		message := SocketMessage{"replicate_value", "Value Replication", []string{key, value}}
		conn.WriteMessage(websocket.TextMessage, message.toString())
	}
}

func (socket *ClusterSocket) ReplicateAllValues(conn *websocket.Conn) {
	if conn == nil {
		return
	}

	message := SocketMessage{"full_replication", "Replicate All Values", socket.database.GetAllValues()}
	conn.WriteMessage(websocket.TextMessage, message.toString())
}

func (socket *ClusterSocket) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := socket.socket.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not establish WebSocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	socket.connections = append(socket.connections, conn)
	opened := true

	conn.SetCloseHandler(func(code int, text string) error {
		opened = false
		fmt.Printf("WebSocket closed: %v (%s)\n", code, text)
		return nil
	})

	for opened {
		err := socket.handleNewMessage(conn)

		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				opened = false
			}
		}
	}
}

func (socket *ClusterSocket) handleNewMessage(conn *websocket.Conn) error {

	if conn == nil {
		return errors.New("nil connection")
	}

	messageType, p, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	switch messageType {
	case websocket.TextMessage:
		{
			var message SocketMessage

			err := json.Unmarshal(p, &message)
			if err == nil && message.EType != "" {
				socket.processMessage(message, conn)
				break
			}

			var response = SocketMessage{"exception", "Unsupported Message Type", nil}
			err = conn.WriteMessage(websocket.TextMessage, response.toString())
		}

	}

	return nil
}

func (socket *ClusterSocket) ConnectToNode(port int) {
	addr := fmt.Sprintf("ws://127.0.0.1:%v/ws", port)

	u, err := url.Parse(addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	socket.connections = append(socket.connections, conn)

	eventMessage := SocketMessage{"connection", "New Cluster Connected", nil}

	err = conn.WriteMessage(websocket.TextMessage, eventMessage.toString())

	if err != nil {
		fmt.Println(err)
		return
	}

	opened := true

	conn.SetCloseHandler(func(code int, text string) error {
		opened = false
		fmt.Printf("WebSocket closed: %v (%s)\n", code, text)
		return nil
	})

	for opened {
		err := socket.handleNewMessage(conn)

		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				opened = false
			}
		}
	}
}
