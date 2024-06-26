package lib

import (
	constants "dbdaddy/const"
	"dbdaddy/db"
	"fmt"

	"github.com/spf13/viper"
)

func PingDB() error {
	if db.DB == nil {
		return fmt.Errorf("error connecting to database, db object is <nil>")
	}

	err := db.DB.Ping()
	if err != nil {
		return fmt.Errorf("error occured while pinging database: %s", err.Error())
	}

	return nil
}

func SwitchDB(v *viper.Viper, dbname string, fn func() error) error {
	defer db.ConnectDb(viper.GetViper(), constants.SelfDbName)

	_, err := db.ConnectDb(v, dbname)
	if err != nil {
		return err
	}

	err = fn()
	if err != nil {
		return err
	}

	return nil
}
