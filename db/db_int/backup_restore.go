package db_int

import (
	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/db/pg"
	"github.com/fossmedaddy/dbdaddy/errs"
	"github.com/fossmedaddy/dbdaddy/globals"
)

func DumpDb(outputFilePath string, dbname string, onlySchema bool) error {
	switch globals.CurrentConnConfig.Driver {
	case constants.DbDriverPostgres:
		return pg.DumpDb(outputFilePath, globals.CurrentConnConfig, onlySchema)
	default:
		return errs.ErrUnsupportedDriver
	}
}

func RestoreDb(dbname string, dumpFilePath string, override bool) error {
	switch globals.CurrentConnConfig.Driver {
	case constants.DbDriverPostgres:
		return pg.RestoreDb(dbname, dumpFilePath, override)
	default:
		return errs.ErrUnsupportedDriver
	}
}
