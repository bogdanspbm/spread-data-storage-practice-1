package utils

import (
	"fmt"
	"net/http"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/objects"
	"spread-data-storage-practice-1/src/utils/requests"
)

type Server struct {
	DatabaseAdapter    *adapters.DatabaseAdapter
	TransactionManager *objects.TransactionManager
}

func CreateServer(databaseAdapter *adapters.DatabaseAdapter, manager *objects.TransactionManager) *Server {
	webServer := &Server{DatabaseAdapter: databaseAdapter, TransactionManager: manager}
	webServer.prepare()
	return webServer
}

func (server *Server) prepare() {
	// Store Handler
	storeServer := requests.CreateStoreServer(server.DatabaseAdapter, server.TransactionManager)
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
