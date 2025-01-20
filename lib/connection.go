package lib

import (
	"fmt"

	"github.com/fossmedaddy/dbdaddy/db"
	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/types"
)

func PingDB() error {
	if globals.DB == nil {
		return fmt.Errorf("error connecting to database, db object is <nil>")
	}

	err := globals.DB.Ping()
	if err != nil {
		return fmt.Errorf("error occured while pinging database: %s", err.Error())
	}

	return nil
}

func TmpSwitchConn(connConfig types.ConnConfig, fn func() error) error {
	defer (func() {
		if globals.DB != nil {
			db.ConnectDb(globals.CurrentConnConfig)
		}
	})()

	_, err := db.ConnectDb(connConfig)
	if err != nil {
		return err
	}

	return fn()
}

// takes in the currently active connection config to switch back to, dbname and
// function to run when connected to the "dbname" database.
// error can be returned if callback function errors or there is a connection error to the DB
func TmpSwitchDB(dbname string, fn func() error) error {
	connConfig := globals.CurrentConnConfig
	connConfig.Database = dbname

	return TmpSwitchConn(connConfig, fn)
}

func TmpSwitchToShadowDB(fn func() error) error {
	shadowConnConfig := libUtils.GetShadowConnConfig()

	if globals.CliConfig.ShadowConnConfig != nil {
		return TmpSwitchConn(shadowConnConfig, fn)
	}

	shadowDbName, err := CreateShadowDB()
	if err != nil {
		return err
	}

	defer db_int.DeleteDb(shadowDbName)
	return TmpSwitchDB(shadowDbName, fn)

}
