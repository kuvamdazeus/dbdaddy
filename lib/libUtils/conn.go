package libUtils

import (
	"fmt"
	"slices"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/errs"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/types"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/maps"
)

func GetConnConfigFromUri(uri string) (types.ConnConfig, error) {
	dbConfig, uriErr := pgx.ParseConfig(uri)
	if uriErr == nil {
		return types.ConnConfig{
			User:       dbConfig.User,
			Password:   dbConfig.Password,
			Host:       dbConfig.Host,
			Port:       fmt.Sprint(dbConfig.Port),
			Database:   dbConfig.Database,
			Params:     dbConfig.RuntimeParams,
			Driver:     constants.DbDriverPostgres,
			ConnString: dbConfig.ConnString(),
		}, nil
	}

	// try parsing for other database drivers (ONLY SUPPORTED ONES)

	return types.ConnConfig{}, errs.ErrUnsupportedDriver
}

func GetShadowConnConfig() types.ConnConfig {
	if globals.CliConfig.ShadowConnConfig != nil {
		return *globals.CliConfig.ShadowConnConfig
	} else {
		return globals.CliConfig.MainConnConfig
	}
}

func GetRemoteConnConfig(cliConfig *types.CliConfig, originKey string) (types.ConnConfig, error) {
	var connConfig types.ConnConfig

	originMKeys := maps.Keys(cliConfig.Origins)

	if !slices.Contains(originMKeys, originKey) {
		return connConfig, fmt.Errorf("remote origin with name '%s' was not found!", originKey)
	}

	connConfig = cliConfig.Origins[originKey]
	return connConfig, nil
}
