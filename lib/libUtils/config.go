package libUtils

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/types"
)

type jsonCliConfig struct {
	DatabaseUrl       string            `json:"database_url"`
	ShadowDatabaseUrl string            `json:"shadow_database_url,omitempty"`
	Origins           map[string]string `json:"origins"`
}

// resolve the env var in config file from either process env or .env file in the config dir
//
// if no env var was found, will return the original value
func resolveConfigEnvVar(k string, vars map[string]string) string {
	if strings.HasPrefix(k, "$") {
		_k := k[1:]
		if v, found := os.LookupEnv(_k); found {
			return v
		}

		if vars[_k] != "" {
			return vars[_k]
		}

		fmt.Println(fmt.Sprintf("WARNING: env var '%s' unused!", k))
	}

	return k
}

func populateJsonCliConfigVars(jsonConfig *jsonCliConfig, vars map[string]string) {
	jsonConfig.DatabaseUrl = resolveConfigEnvVar(jsonConfig.DatabaseUrl, vars)

	for k, v := range jsonConfig.Origins {
		jsonConfig.Origins[k] = resolveConfigEnvVar(v, vars)
	}
}

func validateJsonCliConfig(jsonConfig jsonCliConfig) (types.CliConfig, error) {
	cliConfig := types.NewCliConfig()

	if cc, err := GetConnConfigFromUri(jsonConfig.DatabaseUrl); err != nil {
		return cliConfig, err
	} else {
		cliConfig.MainConnConfig = cc
	}

	if len(jsonConfig.ShadowDatabaseUrl) > 0 {
		if cc, err := GetConnConfigFromUri(jsonConfig.ShadowDatabaseUrl); err != nil {
			return cliConfig, err
		} else {
			cliConfig.ShadowConnConfig = &cc
		}
	}

	for name, uri := range jsonConfig.Origins {
		if cc, err := GetConnConfigFromUri(uri); err != nil {
			return cliConfig, err
		} else {
			cliConfig.Origins[name] = cc
		}
	}

	return cliConfig, nil
}

func getEnvVarOriginKey(originName string) string {
	return fmt.Sprintf("%s%s", constants.EnvVarsOriginKeyPrefix, originName)
}

func ReadCliConfig(configDirPath string, noStateFile bool) (types.CliConfig, error) {
	configFilePath := path.Join(configDirPath, constants.SelfConfigFileName)
	configVarsFilePath := path.Join(configDirPath, constants.SelfEnvVarsFileName)

	EMPTY_CLI_CONFIG := types.NewCliConfig()

	b, readErr := os.ReadFile(configFilePath)
	if readErr != nil {
		return EMPTY_CLI_CONFIG, readErr
	}

	jsonConfig := jsonCliConfig{}
	if err := json.Unmarshal(b, &jsonConfig); err != nil {
		return EMPTY_CLI_CONFIG, err
	}

	vars, varsErr := GetDotenv(configVarsFilePath)
	if varsErr != nil {
		return EMPTY_CLI_CONFIG, varsErr
	}
	populateJsonCliConfigVars(&jsonConfig, vars)

	var cliConfig types.CliConfig
	if cc, err := validateJsonCliConfig(jsonConfig); err != nil {
		return cc, err
	} else {
		cliConfig = cc
	}

	cliConfig.EnvVars = vars

	if noStateFile {
		cliConfig.State = types.CliState{
			CurrentBranch: cliConfig.MainConnConfig.Database,
		}
	} else {
		state, stateErr := ReadCliState(configDirPath)
		if stateErr != nil {
			return EMPTY_CLI_CONFIG, stateErr
		}
		cliConfig.State = state
	}

	return cliConfig, nil
}

// not to be used from outside lib
func WriteCliConfig(cliConfig types.CliConfig, configDirPath string, useEnvVars bool, noStateFile bool) error {
	configFilePath := path.Join(configDirPath, constants.SelfConfigFileName)

	dbUrl := cliConfig.MainConnConfig.ConnString
	shadowDbUrl := ""
	if cliConfig.ShadowConnConfig != nil {
		shadowDbUrl = cliConfig.ShadowConnConfig.ConnString
	}
	origins := map[string]string{}

	if useEnvVars {
		if cliConfig.EnvVars[constants.EnvVarsDatabaseUrlKey] != "" {
			dbUrl = "$" + constants.EnvVarsDatabaseUrlKey
		}

		if shadowDbUrl != "" {
			if cliConfig.EnvVars[constants.EnvVarsShadowDatabaseUrlKey] != "" && useEnvVars {
				shadowDbUrl = "$" + constants.EnvVarsShadowDatabaseUrlKey
			}
		}

		varPairs := [][2]string{}
		for k, v := range cliConfig.EnvVars {
			varPairs = append(varPairs, [2]string{k, v})
		}
		for name, originConnConfig := range cliConfig.Origins {
			findI := slices.IndexFunc(varPairs, func(pair [2]string) bool {
				if pair[1] == originConnConfig.ConnString {
					return true
				}

				return false
			})
			if findI >= 0 && useEnvVars {
				origins[name] = "$" + varPairs[findI][0]
			} else {
				origins[name] = originConnConfig.ConnString
			}
		}
	} else {
		// populate origins without env vars
		for name, originConnConfig := range cliConfig.Origins {
			origins[name] = originConnConfig.ConnString
		}
	}

	jsonConfig := jsonCliConfig{
		DatabaseUrl:       dbUrl,
		ShadowDatabaseUrl: shadowDbUrl,
		Origins:           origins,
	}

	b, marshalErr := json.MarshalIndent(jsonConfig, "", "    ")
	if marshalErr != nil {
		return marshalErr
	}

	if err := os.WriteFile(configFilePath, b, 0666); err != nil {
		return err
	}

	envVarsFilePath := path.Join(configDirPath, constants.SelfEnvVarsFileName)
	if err := SetDotenv(envVarsFilePath, cliConfig.EnvVars); err != nil {
		return err
	}

	if !noStateFile {
		if err := WriteCliState(configDirPath, cliConfig.State); err != nil {
			return err
		}
	}

	return nil
}
