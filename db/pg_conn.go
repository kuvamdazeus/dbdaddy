package db

import (
	"fmt"
	"strings"

	"github.com/fossmedaddy/dbdaddy/types"
)

func GetPgConnUriFromConnConfig(connConfig types.ConnConfig) string {
	urlParams := ""
	for key, val := range connConfig.Params {
		if len(urlParams) == 0 {
			urlParams += "?"
		}

		urlParams += fmt.Sprintf("%s=%s&", key, val)
	}
	urlParams = strings.TrimRight(urlParams, "&")

	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s%s",
		connConfig.User,
		connConfig.Password,
		connConfig.Host,
		connConfig.Port,
		connConfig.Database,
		urlParams,
	)
}
