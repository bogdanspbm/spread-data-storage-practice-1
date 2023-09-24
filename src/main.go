package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
	"spread-data-storage-practice-1/src/utils"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/objects"
)

func main() {
	database, err := sql.Open("sqlite3", "database.sqlite")

	if err != nil {
		return
	}

	defer database.Close()

	adapter := adapters.CreateDatabaseAdapter(database)
	manager := objects.CreateTransactionManager()
	server := utils.CreateServer(adapter, manager)
	server.Start(3000)
}
