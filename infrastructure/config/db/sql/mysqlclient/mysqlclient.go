package mysqlclient

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// New creates a connection from a data source name
func New(dns string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dns)
	if err != nil {
		return nil, fmt.Errorf("error opening connection: %w", err)
	}

	return db, nil
}
