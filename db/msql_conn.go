package db

import "github.com/fossmedaddy/dbdaddy/types"

func GetMysqlConnUriFromViper(cliConfig *types.CliConfig, dbname string) string {
	// return fmt.Sprintf(
	// 	`%s:%s@tcp(%s:%s)/%s?multiStatements=true`,
	// 	v.GetString(constants.DbConfigUserKey),
	// 	v.GetString(constants.DbConfigPassKey),
	// 	v.GetString(constants.DbConfigHostKey),
	// 	v.GetString(constants.DbConfigPortKey),
	// 	dbname,
	// )
	return ""
}
