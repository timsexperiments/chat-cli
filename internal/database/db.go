package database

import (
	"database/sql"
	"fmt"

	"github.com/timsexperiments/chat-cli/internal/constants"
)

type DB struct {
	sql     *sql.DB
	queries *queryCache
}

func CreateDB(sql *sql.DB) *DB {
	db := &DB{sql: sql, queries: newQueryCache()}
	_, err := db.Exec(constants.INIT_QUERY)
	if err != nil {
		panic(fmt.Errorf("unable to initialize database: %w", err))
	}
	return db
}

func (db *DB) Exec(queryName string, args ...any) (sql.Result, error) {
	query, err := db.queries.GetQuery(queryName)
	if err != nil {
		return nil, fmt.Errorf("query %s not found: %w", queryName, err)
	}
	return db.sql.Exec(query, args...)
}

func (db *DB) Query(queryName string, args ...any) (*sql.Rows, error) {
	query, err := db.queries.GetQuery(queryName)
	if err != nil {
		return nil, fmt.Errorf("query %s not found: %w", queryName, err)
	}
	return db.sql.Query(query, args...)
}
