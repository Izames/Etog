package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"path/filepath"
)

func main() {
	var connectionString, migrationPath, migrationTable string
	var force, version, steps int
	var down bool
	flag.StringVar(&connectionString, "conn-str", "", "Connection string")
	flag.StringVar(&migrationPath, "path", "", "Migration Path")
	flag.StringVar(&migrationTable, "table", "migrTable", "Migration Table")
	flag.IntVar(&force, "force", -1, "Force migration")
	flag.IntVar(&steps, "steps", 1, "Number of steps")
	flag.BoolVar(&down, "down", false, "Down migration")
	flag.IntVar(&version, "version", -1, "Version")
	flag.Parse()

	driver, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	instance, err := postgres.WithInstance(driver, &postgres.Config{
		MigrationsTable: migrationTable,
	})
	if err != nil {
		log.Fatal(err)
	}

	absPath, err := filepath.Abs(migrationPath)
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", filepath.ToSlash(absPath)), "postgres", instance)
	if err != nil {
		log.Fatal(err)
	}

	if force >= 0 {
		err = m.Force(force)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("version is forced ", force)
	}

	if down {
		err = m.Steps(-steps)

		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("version is down ", version)
		return
	}

	if version >= 0 {
		err = m.Migrate(uint(version))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("version is ", version)
		return
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("No change")
	} else if err != nil {
		log.Fatal(err)
	}

	fmt.Println("migrations applied")
}
