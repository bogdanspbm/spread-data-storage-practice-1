package utils

import (
	"fmt"
	"net/http"
	"os"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/objects"
	"spread-data-storage-practice-1/src/utils/requests"
	"strings"
	"time"
)

type Server struct {
	DatabaseAdapter    *adapters.DatabaseAdapter
	TransactionManager *objects.TransactionManager
	RequestJournal     []objects.Request
}

func CreateServer(databaseAdapter *adapters.DatabaseAdapter, manager *objects.TransactionManager) *Server {
	webServer := &Server{DatabaseAdapter: databaseAdapter, TransactionManager: manager}
	webServer.prepare()
	go webServer.StartDumpThread()
	return webServer
}

func (server *Server) prepare() {
	// Store Handler
	storeServer := requests.CreateStoreServer(server.DatabaseAdapter, server.TransactionManager, &server.RequestJournal)
	storeHandler := http.HandlerFunc(storeServer.RequestValue)
	http.Handle("/store/", storeHandler)

	// Bind Ping
	pingHandler := http.HandlerFunc(requests.Ping)
	http.Handle("/", pingHandler)
}

func (server *Server) Start(port int) {
	fmt.Println(fmt.Sprintf("<---- Starting Server at Port: %v", port))
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
