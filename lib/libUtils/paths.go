package libUtils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/fossmedaddy/dbdaddy/constants"
)

func GetAbsolutePathFor(relativePath string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return path.Join(cwd, relativePath), err
}

func DirIsProject(dirpath string) (bool, error) {
	if _, dirname := path.Split(dirpath); dirname == constants.SelfConfigDirName {
		return false, nil
	}

	dirEntries, err := os.ReadDir(dirpath)
	if err != nil {
		return false, err
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.Name() == constants.SelfConfigFileName {
			return true, nil
		}
	}

	return false, nil
}

func CwdIsProject() (string, bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return cwd, false, err
	}

	isProject, projectCheckErr := DirIsProject(cwd)
	return cwd, isProject, projectCheckErr
}

func FindConfigFilePath() (string, error) {
	_, cwdFileErr := os.Stat(GetLocalConfigPath())
	_, globalFileErr := os.Stat(GetGlobalConfigPath())

	if cwd, cwdIsProject, err := CwdIsProject(); err != nil {
		return "", err
	} else if cwdIsProject {
		return path.Join(cwd, constants.SelfConfigFileName), nil
	}

	if os.IsNotExist(cwdFileErr) && os.IsNotExist(globalFileErr) {
		return "", fmt.Errorf(globalFileErr.Error() + "\n" + cwdFileErr.Error())
	}

	if os.IsNotExist(cwdFileErr) {
		return GetGlobalConfigPath(), nil
	} else {
		return GetLocalConfigPath(), nil
	}
}

func FindConfigDirPath() (string, error) {
	configFile, err := FindConfigFilePath()
	if err != nil {
		return "", err
	}

	configDir := path.Dir(configFile)

	return configDir, nil
}

func FindTmpDirPath() (string, error) {
	configDir, err := FindConfigDirPath()
	if err != nil {
		return "", err
	}

	tmpDirPath := path.Join(configDir, constants.TmpDir)
	_, osDirErr := EnsureDirExists(tmpDirPath)
	if osDirErr != nil {
		return "", err
	}

	return tmpDirPath, nil
}

func GetDriverDumpDir(configFilePath string, driver string) string {
	configDirPath := path.Dir(configFilePath)

	dumpDirName := constants.DriverDumpDirNames[driver]

	dumpDirPath := path.Join(configDirPath, dumpDirName)
	EnsureDirExists(dumpDirPath)

	return dumpDirPath
}

func GetGlobalConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(homeDir, constants.SelfConfigDirName, constants.SelfConfigFileName)
}

func GetLocalConfigPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Join(cwd, constants.SelfConfigDirName, constants.SelfConfigFileName)
}

func GetGlobalDirPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(homeDir, constants.SelfConfigDirName)
}
