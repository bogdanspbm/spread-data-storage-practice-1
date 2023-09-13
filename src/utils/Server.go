package utils

import (
	"fmt"
	"net/http"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/requests"
)

type Server struct {
	DatabaseAdapter *adapters.DatabaseAdapter
}

func CreateServer(databaseAdapter *adapters.DatabaseAdapter) *Server {
	webServer := &Server{DatabaseAdapter: databaseAdapter}
	webServer.prepare()
	return webServer
}

func (server *Server) prepare() {
	// Bind Ping
	pingHandler := http.HandlerFunc(requests.Ping)
	http.Handle("/", pingHandler)
}

func (server *Server) Start(port int) {
	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
