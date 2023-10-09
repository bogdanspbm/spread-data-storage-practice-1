package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"spread-data-storage-practice-1/src/utils/adapters"
	"strconv"
	"time"
)

type ClusterSocket struct {
	database         *adapters.DatabaseAdapter
	socket           *websocket.Upgrader
	connections      map[*websocket.Conn]bool
	leaderConnection *websocket.Conn
	status           string
	leaderPort       string
	source           string
	logicTime        int
}

func CreateClusterSocket(port string, database *adapters.DatabaseAdapter) *ClusterSocket {
	socket := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin
		},
	}
	return &ClusterSocket{database: database, socket: &socket, connections: make(map[*websocket.Conn]bool, 4), source: port, logicTime: 0}
}

type SocketMessage struct {
	Source     string   `json:"source"`
	LogicTime  int      `json:"logicTime"`
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

func (socket *ClusterSocket) IsLeader() bool {
	return socket.status == "leader"
}

func (socket *ClusterSocket) GetLeaderPort() string {
	return socket.leaderPort
}

func (socket *ClusterSocket) processMessage(message SocketMessage, conn *websocket.Conn) {

	switch message.EType {
	case "connection":
		socket.status = "leader"
		response := SocketMessage{socket.source, socket.GetLogicTimeInc(), "connection_response", "Connection Accepted", nil}
		conn.WriteMessage(websocket.TextMessage, response.toString())
		socket.ReplicateAllValues(conn)
	case "full_replication":
		for i := 0; i < len(message.EArguments)-1; i += 3 {
			key := message.EArguments[i]
			value := message.EArguments[i+1]
			versionStr := message.EArguments[i+2]
			version, _ := strconv.Atoi(versionStr)

			socket.database.SetVersionValue(key, value, version)
		}
	case "replicate_value":

		if message.LogicTime < socket.logicTime {
			break
		}

		if len(message.EArguments) < 2 {
			response := SocketMessage{socket.source, socket.GetLogicTimeInc(), "replicate_value_response", "Bad Arguments", nil}
			conn.WriteMessage(websocket.TextMessage, response.toString())
			break
		}

		key := message.EArguments[0]
		value := message.EArguments[1]

		socket.database.SetVersionValue(key, value, message.LogicTime)
	}

	if message.LogicTime >= socket.logicTime {
		socket.logicTime = message.LogicTime + 1
	}
}

func (socket *ClusterSocket) ReplicateValue(key string, value string) {
	for conn, opened := range socket.connections {
		if conn == nil || !opened {
			continue
		}
		message := SocketMessage{socket.source, socket.GetLogicTimeInc(), "replicate_value", "Value Replication", []string{key, value}}
		conn.WriteMessage(websocket.TextMessage, message.toString())
	}
}

func (socket *ClusterSocket) ReplicateAllValues(conn *websocket.Conn) {
	if conn == nil {
		return
	}

	message := SocketMessage{socket.source, socket.GetLogicTimeInc(), "full_replication", "Replicate All Values", socket.database.GetAllValues()}
	conn.WriteMessage(websocket.TextMessage, message.toString())
}

func (socket *ClusterSocket) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := socket.socket.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not establish WebSocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	socket.connections[conn] = true

	conn.SetCloseHandler(func(code int, text string) error {
		socket.connections[conn] = false
		fmt.Printf("WebSocket closed: %v (%s)\n", code, text)
		return nil
	})

	counter := 0

	for socket.connections[conn] {
		err := socket.handleNewMessage(conn)

		if err != nil {
			counter += 1
		} else {
			counter = 0
		}

		if counter > 10 {
			fmt.Println("Close Websocket Connection")
			socket.connections[conn] = false
		}
	}
}

func (socket *ClusterSocket) GetLogicTimeInc() int {
	socket.logicTime += 1
	return socket.logicTime
}

func (socket *ClusterSocket) GetLogicTime() int {
	return socket.logicTime
}

func (socket *ClusterSocket) SetStatus(status string) {
	socket.status = status
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

			var response = SocketMessage{socket.source, socket.GetLogicTimeInc(), "exception", "Unsupported Message Type", nil}
			err = conn.WriteMessage(websocket.TextMessage, response.toString())
		}

	}

	return nil
}

func (socket *ClusterSocket) pingLeader() {
	counter := 0

	for {
		if socket.leaderConnection != nil {
			err := socket.leaderConnection.WriteMessage(websocket.TextMessage, []byte("{}"))

			if err != nil {
				counter += 1
			} else {
				fmt.Println("Success ping leader")
				counter = 0
			}

			if counter >= 4 {
				counter = 0
				socket.connections[socket.leaderConnection] = false
				socket.leaderConnection.Close()
				socket.leaderConnection = nil
				socket.status = "leader"
				socket.leaderPort = socket.source
				fmt.Println("Can't reach leader. Become a leader")
			}

		}
		time.Sleep(10 * time.Second)
	}
}

func (socket *ClusterSocket) ConnectToNode(port int) {
	addr := fmt.Sprintf("ws://127.0.0.1:%v/ws", port)

	go socket.pingLeader()

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

	socket.status = "follower"
	socket.leaderPort = fmt.Sprintf("%v", port)

	socket.connections[conn] = true
	socket.leaderConnection = conn

	eventMessage := SocketMessage{socket.source, socket.GetLogicTimeInc(), "connection", "New Cluster Connected", nil}

	err = conn.WriteMessage(websocket.TextMessage, eventMessage.toString())

	if err != nil {
		fmt.Println(err)
		return
	}

	conn.SetCloseHandler(func(code int, text string) error {
		socket.connections[conn] = false
		fmt.Printf("WebSocket closed: %v (%s)\n", code, text)
		return nil
	})

	counter := 0

	for socket.connections[conn] {
		err := socket.handleNewMessage(conn)

		if err != nil {
			counter += 1
		} else {
			counter = 0
		}

		if counter > 10 {
			fmt.Println("Close Websocket Connection")
			socket.connections[conn] = false
		}
	}
}
