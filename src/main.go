package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
	"spread-data-storage-practice-1/src/utils"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/objects"
	"spread-data-storage-practice-1/src/utils/ports"
)

var startPort = 3000

func main() {
	database, err := sql.Open("sqlite3", "database.sqlite")

	if err != nil {
		return
	}

	defer database.Close()

	adapter := adapters.CreateDatabaseAdapter(database)
	manager := objects.CreateTransactionManager(adapter)
	server := utils.CreateServer(adapter, manager)

	port := ports.FindAvailablePort(startPort)
	server.Start(port)
}
