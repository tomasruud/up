# Up

A simple, zero dependency, database agnostic, migration library for Go projects.

Up, as the name suggests, only supports migrations in a single direction. This means that a migration can never be rolled back. If you need to revert a change, you have to append a new migration that reverts your previous change. If you need to be able to migrate down, this library is probably not for you.

Check out the complete docs at https://go.dev/pkg/github.com/tomasruud/up

## Usage
### Minimal
Below is a minimal example on how migrations for a SQLite database could be done.

```go
package main

import (
	"database/sql"
	"log"

	"github.com/tomasruud/up"
	_ "modernc.org/sqlite"
)

func main() {
	db, _ := sql.Open("sqlite", ":memory:")

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
		},
	}

	if err := migrator.Migrate(db); err != nil {
		log.Fatalf("Unable to run migrations: %v", err)
	}

	log.Println("Migrations done")
}

```

### Full
Below is a full example on how this library could be used, including using callback hooks for doing logging. This example uses the provided `StateStore` for SQLite, but you are free to make your own implementation that suit your needs.

```go
package main

import (
	"database/sql"
	"log"

	"github.com/tomasruud/up"
	_ "modernc.org/sqlite"
)

func main() {
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
```