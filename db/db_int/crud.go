package db_int

import (
	constants "dbdaddy/const"
	"dbdaddy/db/msql"
	"dbdaddy/db/pg"
	"dbdaddy/errs"
	"dbdaddy/types"

	"github.com/spf13/viper"
)

// this is a central place for calling driver-specific internal functions

func GetExistingDbs() ([]string, error) {
	driver := viper.GetString(constants.DbConfigDriverKey)
	switch driver {
	case constants.DbDriverPostgres:
		return pg.GetExistingDbs()
	case constants.DbDriverMySQL:
		return msql.GetExistingDbs()
	default:
		return []string{}, errs.ErrUnsupportedDriver
	}
}

func DisconnectAllUsers(dbname string) error {
	driver := viper.GetString(constants.DbConfigDriverKey)
	if driver == constants.DbDriverPostgres {
		return pg.DisconnectAllUsers(dbname)
	} else {
		panic("")
	}
}

func DbExists(dbname string) bool {
	driver := viper.GetString(constants.DbConfigDriverKey)
	if driver == constants.DbDriverPostgres {
		return pg.DbExists(dbname)
	} else {
		panic("")
	}
}

func NewDbFromOriginal(originalDbName string, newDbName string) error {
	driver := viper.GetString(constants.DbConfigDriverKey)
	if driver == constants.DbDriverPostgres {
		return pg.NewDbFromOriginal(originalDbName, newDbName)
	} else {
		panic("")
	}
}

func CreateDb(dbname string) error {
	driver := viper.GetString(constants.DbConfigDriverKey)
	if driver == constants.DbDriverPostgres {
		return pg.CreateDb(dbname)
	} else {
		panic("")
	}
}

func DeleteDb(dbname string) error {
	driver := viper.GetString(constants.DbConfigDriverKey)
	if driver == constants.DbDriverPostgres {
		return pg.DeleteDb(dbname)
	} else {
		panic("")
	}
}

func ListTablesInDb() ([]types.Table, error) {
	driver := viper.GetString(constants.DbConfigDriverKey)
	if driver == constants.DbDriverPostgres {
		return pg.ListTablesInDb()
	} else {
		panic("")
	}
}

func GetTableSchema(schema string, tablename string) (types.TableSchema, error) {
	driver := viper.GetString(constants.DbConfigDriverKey)
	if driver == constants.DbDriverPostgres {
		return pg.GetTableSchema(schema, tablename)
	} else {
		panic("")
	}
}
