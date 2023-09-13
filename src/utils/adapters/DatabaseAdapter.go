package adapters

import "database/sql"

type DatabaseAdapter struct {
	connection *sql.DB
}

func CreateDatabaseAdapter(connection *sql.DB) *DatabaseAdapter {
	return &DatabaseAdapter{connection: connection}
}
