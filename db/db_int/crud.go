package db_int

import (
	constants "dbdaddy/const"
	"dbdaddy/db/pg"
	"dbdaddy/types"

	"github.com/spf13/viper"
)

// this is a central place for calling driver-specific internal functions

func GetExistingDbs() ([]string, error) {
	driver := viper.GetString(constants.DbConfigDriverKey)
	if driver == constants.DbDriverPostgres {
		return pg.GetExistingDbs()
	} else {
		panic("")
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

func ListTablesInDb(dbname string) ([]types.Table, error) {
	driver := viper.GetString(constants.DbConfigDriverKey)
	if driver == constants.DbDriverPostgres {
		return pg.ListTablesInDb(dbname)
	} else {
		panic("")
	}
}

func GetTableSchema(dbname string, schema string, tablename string) (types.TableSchema, error) {
	driver := viper.GetString(constants.DbConfigDriverKey)
	if driver == constants.DbDriverPostgres {
		return pg.GetTableSchema(dbname, schema, tablename)
	} else {
		panic("")
	}
}
