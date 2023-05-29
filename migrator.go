package up

import (
	"database/sql"
	"errors"
	"fmt"
)

type Migration func(*sql.Tx) error

// StateStore is an interface that should be implemented to store the current migration index.
type StateStore interface {
	// Prepare should prepare the state store for use.
	Prepare(*sql.Tx) error

	// Get should return the index of the last migration that was run, if no migrations has run nil should be returned.
	Get(*sql.Tx) (*int, error)

	// Set should set the index of the last migration that was run.
	Set(*sql.Tx, int) error
}

type Migrator struct {
	StateStore StateStore
	Migrations []Migration

	OnSetupComplete func(start int, last *int)
	OnMigrationDone func(current int, start int)
}

// Migrate runs all migrations that have not been run yet. If a migration fails, the transaction is rolled back and the
// error is returned.
func (m Migrator) Migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := m.migrateTx(tx); err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			return fmt.Errorf("migration failed, and failed to rollback transaction: %w", errors.Join(err2, err))
		}

		return fmt.Errorf("migration failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m Migrator) migrateTx(tx *sql.Tx) error {
	if err := m.store().Prepare(tx); err != nil {
		return fmt.Errorf("failed to prepare state store: %w", err)
	}

	last, err := m.store().Get(tx)
	if err != nil {
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	start := 0
	if last != nil {
		start = *last + 1
	}

	if start >= len(m.Migrations) {
		return fmt.Errorf("last migration index (%d) is out of bounds (<%d)", start, len(m.Migrations))
	}

	if m.OnSetupComplete != nil {
		m.OnSetupComplete(start, last)
	}

	for i, migration := range m.Migrations[start:] {
		if err := migration(tx); err != nil {
			return fmt.Errorf("failed to run migration: %v", err)
		}

		if err := m.store().Set(tx, start+i); err != nil {
			return fmt.Errorf("failed to update migration state: %v", err)
		}

		if m.OnMigrationDone != nil {
			m.OnMigrationDone(i, start)
		}
	}

	return nil
}

func (m Migrator) store() StateStore {
	if m.StateStore == nil {
		return &NilStore{}
	}

	return m.StateStore
}
