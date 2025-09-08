package migrator

import "github.com/golang-migrate/migrate/source"

type Migrator struct {
	srcDriver source.Driver
}

//func NewMigrator(srcDriver source.Driver) *Migrator {
//
//}
