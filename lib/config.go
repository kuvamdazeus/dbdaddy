package lib

import (
	"fmt"
	"path"
	"slices"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/types"
	"golang.org/x/exp/maps"
)

func IsFirstTimeUser() bool {
	_, err := libUtils.FindConfigFilePath()
	return err != nil
}

// use InitCliConfig() if no default config is there
func InitConfigFile(cliConfig types.CliConfig, configDirPath string, useEnvVars bool) error {
	globals.CliConfig = cliConfig

	if _, err := libUtils.EnsureDirExists(configDirPath); err != nil {
		return err
	}

	if err := libUtils.WriteCliConfig(globals.CliConfig, configDirPath, useEnvVars); err != nil {
		return err
	}

	return nil
}

func InitCliConfig() types.CliConfig {
	defaultConnConfig := types.NewDefaultPgConnConfig()

	vars := types.CliConfigVars{}
	vars[constants.EnvVarsDatabaseUrlKey] = defaultConnConfig.ConnString

	cliConfig := types.CliConfig{
		MainConnConfig: defaultConnConfig,
		EnvVars:        vars,
		Origins:        map[string]types.ConnConfig{},
		State: types.CliState{
			CurrentBranch: defaultConnConfig.Database,
		},
	}

	return cliConfig
}

// for global config updates and global config related logic, use WriteGlobalConfig
//
// check for cwdIsProject for using `useEnvVars` on your own
//
// FOR GLOBAL CONFIG `useEnvVars` SHALL NOT BE TRUE
func WriteConfig(cliConfig types.CliConfig, configDirPath string, useEnvVars bool) error {
	if useEnvVars {
		if cliConfig.EnvVars[constants.EnvVarsDatabaseUrlKey] != "" {
			cliConfig.EnvVars[constants.EnvVarsDatabaseUrlKey] = cliConfig.MainConnConfig.ConnString
		}

		if err := libUtils.SetDotenv(
			path.Join(configDirPath, constants.SelfEnvVarsFileName),
			cliConfig.EnvVars,
		); err != nil {
			return err
		}
	}

	if err := libUtils.WriteCliConfig(cliConfig, configDirPath, useEnvVars); err != nil {
		return err
	}

	return nil
}

func ReadConfig(configDirPath string, saveInGlobalConfig bool) (types.CliConfig, error) {
	var cliConfig types.CliConfig
	if cc, err := libUtils.ReadCliConfig(configDirPath); err != nil {
		return cliConfig, err
	} else {
		cliConfig = cc
	}

	if saveInGlobalConfig {
		globals.CliConfig = cliConfig
	}

	return cliConfig, nil
}

func EnsureSupportedDbDriver() error {
	if !slices.Contains(constants.SupportedDrivers, globals.CliConfig.MainConnConfig.Driver) {
		return fmt.Errorf(
			"unsupported database driver '%s' supported drivers are: %v",
			globals.CurrentConnConfig.Driver,
			constants.SupportedDrivers,
		)
	}

	return nil
}

// uses cmd remote flag options to choose between main connection config and remote origins
func GetSuitableConnConfig(cliConfig *types.CliConfig, remoteFlag bool, remoteNameFlag string) (types.ConnConfig, error) {
	var connConfig types.ConnConfig
	if remoteNameFlag != "" {
		if cc, err := libUtils.GetRemoteConnConfig(cliConfig, remoteNameFlag); err != nil {
			return connConfig, fmt.Errorf("unexpected error occured while fetching remote connection config!\n%s", err)
		} else {
			connConfig = cc
		}
	} else if remoteFlag {
		originMKeys := maps.Keys(cliConfig.Origins)

		if len(originMKeys) > 1 {
			return connConfig, fmt.Errorf("found more than 1 remote origins, please provide a remote origin name.")
		} else if len(originMKeys) == 0 {
			return connConfig, fmt.Errorf("no remote origin found, please add a remote origin database.")
		}

		connConfig = cliConfig.Origins[originMKeys[0]]
	} else {
		connConfig = globals.CliConfig.MainConnConfig
		connConfig.Database = globals.CliConfig.State.CurrentBranch
	}

	return connConfig, nil
}
