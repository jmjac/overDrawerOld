package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func Open(user, password, dbName, host string, port int) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	return db, err
}
