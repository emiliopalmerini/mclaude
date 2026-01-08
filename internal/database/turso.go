package database

import (
	"database/sql"

	_ "github.com/tursodatabase/go-libsql"
)

func NewTurso(databaseURL, authToken string) (*sql.DB, error) {
	connStr := databaseURL + "?authToken=" + authToken
	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
