package db

import (
	"database/sql"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/errs"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/types"

	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openConn(driverName string, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		fmt.Println(fmt.Sprintf("WARNING: ping failed for '%s',", dataSourceName), err)
		return nil, errs.ErrDbPingFailed
	}

	return db, nil
}

func ConnectSelfDb(connConfig types.ConnConfig) (*sql.DB, error) {
	var fnLocalDb *sql.DB

	selfConnConfig := connConfig
	selfConnConfig.Database = constants.SelfDbName

	db, err := ConnectDb(selfConnConfig)
	if errors.Is(err, errs.ErrUnsupportedDriver) {
		return db, err
	} else if err != nil {
		userDb, userErr := ConnectDb(connConfig)
		if userErr != nil {
			return nil, fmt.Errorf("error connecting to your database!\n" + err.Error())
		}
		defer userDb.Close()

		if _, err := userDb.Query(fmt.Sprintf("CREATE DATABASE %s", constants.SelfDbName)); err != nil {
			return nil, fmt.Errorf("could not create database, please check your connection or permissions!\n" + err.Error())
		}

		selfDb, selfErr := ConnectDb(selfConnConfig)
		if selfErr != nil {
			return nil, fmt.Errorf("unexpected error occured!\n" + selfErr.Error())
		}

		fnLocalDb = selfDb
	} else {
		fnLocalDb = db
	}

	globals.CurrentConnConfig = selfConnConfig
	globals.DB = fnLocalDb

	return globals.DB, nil
}

func ConnectDb(connConfig types.ConnConfig) (*sql.DB, error) {
	var (
		db  *sql.DB
		err error
	)

	switch connConfig.Driver {
	case constants.DbDriverPostgres:
		db, err = openConn("pgx", GetPgConnUriFromConnConfig(connConfig))
	default:
		return db, errs.ErrUnsupportedDriver
	}

	if err != nil {
		return db, err
	}

	globals.CurrentConnConfig = connConfig
	globals.DB = db

	return globals.DB, nil
}
