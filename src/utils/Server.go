package utils

import (
	"fmt"
	"net/http"
	"os"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/objects"
	"spread-data-storage-practice-1/src/utils/ports"
	"spread-data-storage-practice-1/src/utils/requests"
	"spread-data-storage-practice-1/src/utils/websocket"
	"strings"
	"time"
)

type Server struct {
	Port               int
	Websocket          *websocket.ClusterSocket
	DatabaseAdapter    *adapters.DatabaseAdapter
	TransactionManager *objects.TransactionManager
	RequestJournal     []objects.Request
}

func CreateServer(databaseAdapter *adapters.DatabaseAdapter, socket *websocket.ClusterSocket, manager *objects.TransactionManager) *Server {
	webServer := &Server{DatabaseAdapter: databaseAdapter, Websocket: socket, TransactionManager: manager}
	webServer.prepare()
	go webServer.StartDumpThread()
	return webServer
}

func (server *Server) prepare() {
	// Store Handler
	storeServer := requests.CreateStoreServer(server.DatabaseAdapter, server.Websocket, server.TransactionManager, &server.RequestJournal)
	storeHandler := http.HandlerFunc(storeServer.RequestValue)
	http.Handle("/store/", storeHandler)

	// Bind Ping
	pingHandler := http.HandlerFunc(requests.Ping)
	http.Handle("/", pingHandler)

	// Bind Clock
	clockHandler := http.HandlerFunc(storeServer.GetClock)
	http.Handle("/vclock", clockHandler)

	// Bind Test Page
	testPageHandler := http.HandlerFunc(requests.TestPage)
	http.Handle("/test", testPageHandler)

	// Create Socket
	socketHandler := http.HandlerFunc(server.Websocket.Handler)
	http.Handle("/ws", socketHandler)
}

func (server *Server) establishClusterConnection() {
	server.Websocket.ConnectToNode(3000)
}

func (server *Server) Start(port int) {
	fmt.Println(fmt.Sprintf("<---- Starting Server at Port: %v ---->", port))

	server.Port = port

	if port > 3000 && !ports.IsPortOpen(3000) {
		go server.Websocket.ConnectToNode(3000)
	}

	if port == 3000 && !ports.IsPortOpen(3001) {
		go server.Websocket.ConnectToNode(3001)
	} else if port == 3000 {
		server.Websocket.SetStatus("leader")
	}

	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

func (server *Server) StartDumpThread() {
	for {
		time.Sleep(1 * time.Minute)
		server.dumpJournal()
	}
}

func (server *Server) dumpJournal() {
	builder := strings.Builder{}
	fmt.Println(fmt.Sprintf("Start Dump Process: %v", len(server.RequestJournal)))

	for _, request := range server.RequestJournal {
		builder.WriteString(request.Name)
		builder.WriteString("\n")
		builder.WriteString(strings.Join(request.Args, ","))
		builder.WriteString("\n")
	}

	os.WriteFile("dump.belarus", []byte(builder.String()), 0666)
}
