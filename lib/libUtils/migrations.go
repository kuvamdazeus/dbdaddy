package libUtils

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/fossmedaddy/dbdaddy/constants"
)

func GetLocalMigrationsDir(configDirPath, dbname string) string {
	return path.Join(configDirPath, constants.MigDirName, constants.MigLocalDirName, dbname)
}

func GetRemoteMigrationsDir(configDirPath, dbname string) string {
	return path.Join(configDirPath, constants.MigDirName, dbname)
}

// returns "00069__title" or "00069"
func GenerateMigrationId(migrationsDirPath string, title string) (string, error) {
	dirs, err := os.ReadDir(migrationsDirPath)
	if err != nil {
		return "", err
	}

	latestN := 0
	for _, dir := range dirs {
		n, err := strconv.Atoi(dir.Name())
		if err != nil {
			return "", err
		}

		latestN = max(latestN, n+1)
	}

	fileNum := fmt.Sprintf("%05d", latestN)

	fileTitle := strings.ReplaceAll(strings.Trim(strings.ToLower(title), " "), " ", "_")

	if len(fileTitle) > 0 {
		return fileNum + "__" + fileTitle, nil
	}

	return fileNum, nil
}
