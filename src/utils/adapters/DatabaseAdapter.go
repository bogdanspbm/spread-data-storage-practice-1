package adapters

import (
	"database/sql"
	"errors"
	"fmt"
)

type DatabaseAdapter struct {
	connection *sql.DB
	inMem      map[string]VersionValue
}

type Data struct {
	Key   string `json:"key" db:"key"`
	Value string `json:"value" db:"value"`
}

type VersionData struct {
	Key     string `json:"key" db:"key"`
	Value   string `json:"value" db:"value"`
	Version int    `json:"version" db:"version"`
}

type VersionValue struct {
	Value   string `json:"value" db:"value"`
	Version int    `json:"version" db:"version"`
}

func CreateDatabaseAdapter(connection *sql.DB) *DatabaseAdapter {
	return &DatabaseAdapter{connection: connection, inMem: make(map[string]VersionValue)}
}

func (adapter *DatabaseAdapter) GetAllValues() []string {
	output := make([]string, 0)
	for k, v := range adapter.inMem {
		output = append(output, k)
		output = append(output, v.Value)
		output = append(output, fmt.Sprint("%v", v.Version))
	}
	return output
}

func (adapter *DatabaseAdapter) GetValue(key string) (VersionValue, error) {
	/*rows, err := adapter.connection.Query("SELECT * FROM stored_data WHERE key=$1", key)
	defer rows.Close()

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

	return "", errors.New("can't find key")*/

	v, ok := adapter.inMem[key]

	var err error

	if !ok {
		err = errors.New("empty value")
	}

	return v, err
}

func (adapter *DatabaseAdapter) SetValue(key string, value VersionValue) (int64, error) {
	adapter.inMem[key] = value
	return int64(len(adapter.inMem)), nil
	/*query := "INSERT INTO stored_data (key, value)  VALUES($1,$2)  ON CONFLICT(key)  DO UPDATE SET value=excluded.value;"
	res, err := adapter.connection.Exec(query, key, value)

	if err != nil {
		return -1, err
	}

	return res.LastInsertId()*/
}

func (adapter *DatabaseAdapter) SetVersionValue(key string, value string, version int) (int64, error) {
	oldValue, _ := adapter.GetValue(key)

	if oldValue.Version > version {
		return -1, errors.New("database version is bigger")
	}

	adapter.inMem[key] = VersionValue{value, version}
	return int64(len(adapter.inMem)), nil
}
