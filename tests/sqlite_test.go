package tests

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/tomasruud/up"
	_ "modernc.org/sqlite"
)

func TestSQLiteStore(t *testing.T) {
	newTx := func(t *testing.T) *sql.Tx {
		t.Helper()

		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			panic(fmt.Errorf("unable to open database: %w", err))
		}

		tx, err := db.Begin()
		if err != nil {
			panic(fmt.Errorf("unable to begin transaction: %w", err))
		}

		t.Cleanup(func() {
			if err := tx.Rollback(); err != nil {
				panic(fmt.Errorf("unable to rollback transaction: %w", err))
			}

			if err := db.Close(); err != nil {
				panic(fmt.Errorf("unable to close database: %w", err))
			}
		})

		return tx
	}

	t.Run("prepare", func(t *testing.T) {
		t.Run("it creates a migrations table with default name if none given", func(t *testing.T) {
			tx := newTx(t)

			err := up.SQLiteStore{}.Prepare(tx)
			if err != nil {
				t.Errorf("err was not nil: %v", err)
			}

			var got string
			if err := tx.QueryRow(`select sql from sqlite_schema where name = "migrations"`).Scan(&got); err != nil {
				t.Errorf("failed listing table specs: %v", err)
			}

			want := `CREATE TABLE migrations (id int primary key, created text)`
			if got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		})

		t.Run("it creates a migrations table with custom name if given", func(t *testing.T) {
			tx := newTx(t)

			err := up.SQLiteStore{Table: "yeet"}.Prepare(tx)
			if err != nil {
				t.Errorf("err was not nil: %v", err)
			}

			var got string
			if err := tx.QueryRow(`select sql from sqlite_schema where name = "yeet"`).Scan(&got); err != nil {
				t.Errorf("failed listing table specs: %v", err)
			}

			want := `CREATE TABLE yeet (id int primary key, created text)`
			if got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	})

	t.Run("get", func(t *testing.T) {
		t.Run("it returns nil if no migration records exists", func(t *testing.T) {
			tx := newTx(t)

			s := up.SQLiteStore{}
			if err := s.Prepare(tx); err != nil {
				panic(fmt.Errorf("unable to prepare store: %w", err))
			}

			got, err := s.Get(tx)
			if err != nil {
				t.Errorf("err was not nil: %v", err)
			}

			if got != nil {
				t.Errorf("got %v, want nil", got)
			}

			var table string
			if err := tx.QueryRow(`select sql from sqlite_schema where name = "migrations"`).Scan(&table); err != nil {
				t.Errorf("failed listing table specs: %v", err)
			}

			want := `CREATE TABLE migrations (id int primary key, created text)`
			if table != want {
				t.Errorf("got %v, want %v", table, want)
			}
		})

		t.Run("it returns the latest migration if any", func(t *testing.T) {
			tx := newTx(t)

			s := up.SQLiteStore{}
			if err := s.Prepare(tx); err != nil {
				panic(fmt.Errorf("unable to prepare store: %w", err))
			}

			if err := s.Set(tx, 123); err != nil {
				panic(fmt.Errorf("unable to set migration: %w", err))
			}

			got, err := s.Get(tx)
			if err != nil {
				t.Errorf("err was not nil: %v", err)
			}

			if got == nil {
				t.Errorf("got nil, want 123")
			}

			if *got != 123 {
				t.Errorf("got %v, want 123", *got)
			}
		})
	})

	t.Run("set", func(t *testing.T) {
		t.Run("it inserts a new migration record", func(t *testing.T) {
			tx := newTx(t)

			s := up.SQLiteStore{}
			if err := s.Prepare(tx); err != nil {
				panic(fmt.Errorf("unable to prepare store: %w", err))
			}

			err := s.Set(tx, 312)
			if err != nil {
				t.Errorf("err was not nil: %v", err)
			}

			var got *int
			if err := tx.QueryRow(`select id from migrations limit 1`).Scan(&got); err != nil {
				t.Errorf("failed getting migration: %v", err)
			}

			if got == nil {
				t.Errorf("got nil, want 312")
			}

			if *got != 312 {
				t.Errorf("got %v, want 312", *got)
			}
		})
	})
}
