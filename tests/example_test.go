package tests

import (
	"database/sql"
	"log"

	"github.com/tomasruud/up"
	_ "modernc.org/sqlite"
)

func Example() {
	db, _ := sql.Open("sqlite", ":memory:")

	migrations := []up.Migration{
		func(tx *sql.Tx) error {
			_, err := tx.Exec(`create table users(id text primary key)`)
			return err
		},

		func(tx *sql.Tx) error {
			_, err := tx.Exec(`alter table users add column name text`)
			return err
		},
	}

	migrator := up.Migrator{
		StateStore: up.SQLiteStore{},
		Migrations: migrations,

		OnSetupComplete: func(start int, last *int) {
			if last == nil {
				log.Printf("Migrations initialized, no migrations have run")
				return
			}

			log.Printf("Migrations initialized, latest migration is %d", *last)
		},

		OnMigrationDone: func(current int, start int) {
			log.Printf("Migration %d/%d done", current+1+start, len(migrations)-start)
		},
	}

	if err := migrator.Migrate(db); err != nil {
		log.Fatalf("Unable to run migrations: %v", err)
	}

	log.Println("Migrations done")
}
