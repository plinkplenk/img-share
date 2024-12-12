package main

import (
	"context"
	"flag"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"log"
)

const commands = `Commands:
    up                   Migrate the DB to the most recent version available
    up-by-one            Migrate the DB up by 1
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    reset                Roll back all migrations
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with the current timestamp
    fix                  Apply sequential ordering to migrations
    validate             Check migration files without running them
`

var (
	dir      = flag.String("dir", "./migrations/", "directory with migration files")
	dbString = flag.String("db", "postgresql://postgres:postgres@localhost:5432/postgres", "postgres db url")
	help     = flag.Bool("help", false, "execute this command")
	h        = flag.Bool("h", false, "execute this command")
)

func usage() {
	flag.Usage()
	fmt.Print(commands)
}

func main() {
	flag.Parse()
	if *help || *h {
		usage()
		return
	}
	db, err := goose.OpenDBWithDriver("pgx", *dbString)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("goose: migrations command requires at least one argument.")
	}
	command := args[0]

	if err := goose.RunContext(context.Background(), command, db, *dir, args[1:]...); err != nil {
		usage()
		log.Fatalf("goose %v: %v", command, err)
	}
}
