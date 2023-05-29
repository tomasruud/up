package up

import (
	"database/sql"
	"fmt"
	"time"
)

// SQLiteStore is a StateStore implementation for SQLite.
// This is just a basic example, you are free to implement your own custom logic with whatever fields and logic you might need.
type SQLiteStore struct {
	// Table is the name of the table to use for storing the current migration index, defaults to "migrations" if nothing is set.
	Table string

	// Now is a function that returns the current time, defaults to time.Now if nothing is set.
	Now func() time.Time
}

func (s SQLiteStore) Prepare(tx *sql.Tx) error {
	q := fmt.Sprintf("create table if not exists %s (id int primary key, created text)", s.table())
	if _, err := tx.Exec(q); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}
	return nil
}

func (s SQLiteStore) Get(tx *sql.Tx) (*int, error) {
	var current *int
	q := fmt.Sprintf("select max(id) from %s", s.table())
	if err := tx.QueryRow(q).Scan(&current); err != nil {
		return nil, fmt.Errorf("failed to get current migration: %v", err)
	}

	return current, nil
}

func (s SQLiteStore) Set(tx *sql.Tx, i int) error {
	q := fmt.Sprintf("insert into %s (id, created) values (?, ?)", s.table())
	if _, err := tx.Exec(q, i, s.now()); err != nil {
		return fmt.Errorf("failed to insert migration: %v", err)
	}

	return nil
}

func (s SQLiteStore) table() string {
	if s.Table == "" {
		return "migrations"
	}

	return s.Table
}

func (s SQLiteStore) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}

	return time.Now()
}
