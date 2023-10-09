package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
	"spread-data-storage-practice-1/src/utils"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/objects"
	"spread-data-storage-practice-1/src/utils/ports"
	"spread-data-storage-practice-1/src/utils/websocket"
)

var startPort = 3000

func main() {
	database, err := sql.Open("sqlite3", "database.sqlite")

	if err != nil {
		return
	}

	defer database.Close()

	port := ports.FindAvailablePort(startPort)

	adapter := adapters.CreateDatabaseAdapter(database)
	socket := websocket.CreateClusterSocket(fmt.Sprint("%v", port), adapter)
	manager := objects.CreateTransactionManager(adapter, socket)
	server := utils.CreateServer(adapter, socket, manager)

	server.Start(port)
}
