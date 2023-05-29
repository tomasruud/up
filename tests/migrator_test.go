package tests

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/tomasruud/up"
)

func TestMigrator(t *testing.T) {
	newDB := func(t *testing.T) *sql.DB {
		t.Helper()

		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			panic(fmt.Errorf("unable to open database: %w", err))
		}

		t.Cleanup(func() {
			t.Helper()
			if err := db.Close(); err != nil {
				panic(fmt.Errorf("unable to close database: %w", err))
			}
		})

		return db
	}

	t.Run("it migrates stuff", func(t *testing.T) {
		db := newDB(t)

		onSetupComplete := false
		onMigrationDone := 0

		migrator := up.Migrator{
			StateStore: up.SQLiteStore{},
			Migrations: []up.Migration{
				func(tx *sql.Tx) error {
					_, err := tx.Exec(`create table users(id text primary key)`)
					return err
				},

				func(tx *sql.Tx) error {
					_, err := tx.Exec(`alter table users add column name text`)
					return err
				},

				func(tx *sql.Tx) error {
					_, err := tx.Exec(`insert into users(id, name) values('1', 'tomas')`)
					return err
				},
			},
			OnSetupComplete: func(start int, last *int) {
				onSetupComplete = true
			},
			OnMigrationDone: func(current int, start int) {
				onMigrationDone++
			},
		}

		if err := migrator.Migrate(db); err != nil {
			t.Fatalf("unable to migrate: %v", err)
		}

		var got int
		if err := db.QueryRow("select count(*) from migrations").Scan(&got); err != nil {
			t.Fatalf("unable to query migrations: %v", err)
		}

		want := 3
		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}

		if !onSetupComplete {
			t.Errorf("onSetupComplete was not called")
		}

		if onMigrationDone != 3 {
			t.Errorf("onMigrationDone was not called 3 times, got %d", onMigrationDone)
		}
	})

	t.Run("it rolls back if one migration fails", func(t *testing.T) {
		db := newDB(t)

		onMigrationDone := 0

		migrator := up.Migrator{
			StateStore: up.SQLiteStore{},
			Migrations: []up.Migration{
				func(tx *sql.Tx) error {
					_, err := tx.Exec(`create table users(id text primary key)`)
					return err
				},

				func(tx *sql.Tx) error {
					_, err := tx.Exec(`alter table users add column name text`)
					return err
				},

				func(tx *sql.Tx) error {
					return errors.New("oops")
				},
			},
			OnMigrationDone: func(current int, start int) {
				onMigrationDone++
			},
		}

		if err := migrator.Migrate(db); err == nil {
			t.Fatalf("expected an error, got nil")
		}

		var got int
		if err := db.QueryRow(`select count(*) from sqlite_master where type="table" and name="migrations"`).Scan(&got); err != nil {
			t.Fatalf("unable to query migrations: %v", err)
		}

		if got != 0 {
			t.Error("found migrations table, did not expect that")
		}

		var got2 int
		if err := db.QueryRow(`select count(*) from sqlite_master where type="table" and name="users"`).Scan(&got); err != nil {
			t.Fatalf("unable to query users table: %v", err)
		}

		if got2 != 0 {
			t.Error("found users table, did not expect that")
		}

		if onMigrationDone != 2 {
			t.Errorf("onMigrationDone was not called 2 times, got %d", onMigrationDone)
		}
	})

}
