package adapters

import (
	"database/sql"
	"errors"
)

type DatabaseAdapter struct {
	connection *sql.DB
}

type Data struct {
	Key   string `json:"key" db:"key"`
	Value string `json:"value" db:"value"`
}

func CreateDatabaseAdapter(connection *sql.DB) *DatabaseAdapter {
	return &DatabaseAdapter{connection: connection}
}

func (adapter *DatabaseAdapter) GetValue(key string) (string, error) {
	rows, err := adapter.connection.Query("SELECT * FROM stored_data WHERE key=$1", key)

	if err != nil {
		return "", err
	}

	for rows.Next() {
		var key string
		var value string

		err = rows.Scan(&key, &value)

		if err != nil {
			return "", err
		}

		return value, nil
	}

	return "", errors.New("can't find key")
}

func (adapter *DatabaseAdapter) SetValue(key string, value string) (int64, error) {
	query := "INSERT INTO stored_data (key, value)  VALUES($1,$2)  ON CONFLICT(key)  DO UPDATE SET value=excluded.value;"
	res, err := adapter.connection.Exec(query, key, value)

	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}
