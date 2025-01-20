package libUtils

import (
	"encoding/json"
	"os"
	"path"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/types"
)

func ReadCliState(configDirPath string) (types.CliState, error) {
	stateFilePath := path.Join(configDirPath, constants.SelfStateFileName)

	cliState := types.CliState{}

	b, readErr := os.ReadFile(stateFilePath)
	if readErr != nil {
		return cliState, readErr
	}

	if err := json.Unmarshal(b, &cliState); err != nil {
		return cliState, err
	}

	return cliState, nil
}

func WriteCliState(configDirPath string, cliState types.CliState) error {
	stateFilePath := path.Join(configDirPath, constants.SelfStateFileName)

	b, marshalErr := json.MarshalIndent(cliState, "", "    ")
	if marshalErr != nil {
		return marshalErr
	}

	f, fileErr := os.Create(stateFilePath)
	if fileErr != nil {
		return fileErr
	}

	if _, err := f.Write(b); err != nil {
		return err
	}

	return nil
}
