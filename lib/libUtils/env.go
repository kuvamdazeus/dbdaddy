package libUtils

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fossmedaddy/dbdaddy/types"
	"golang.org/x/exp/maps"
)

func errInvalidSegmentsDotenvLineRead(lineNum int, filePath string) error {
	return fmt.Errorf("error at line %d in env file %s: invalid '=' separated segments! need variables defined as 'VAR=VALUE'", lineNum, filePath)
}

// cleans the line from .env file and returns (skipReading, cleanedLine)
func cleanDotenvLine(line string) (bool, string) {
	if strings.HasPrefix(line, "#") {
		return true, line
	}

	line = strings.Trim(line, " ")
	line = strings.ReplaceAll(line, "\"", "")
	line = strings.ReplaceAll(line, "'", "")

	return len(strings.Trim(line, fmt.Sprintln())) == 0, line
}

// returns empty values and nil error if getDotenv fails to read env file due to NOT EXIST error
func getDotenv(filePath string) ([]byte, types.CliConfigVars, error) {
	vars := types.CliConfigVars{}

	fileB, readErr := os.ReadFile(filePath)
	if readErr != nil {
		if errors.Is(readErr, os.ErrNotExist) {
			return []byte{}, vars, nil
		}

		return fileB, vars, readErr
	}

	for i, line := range strings.Split(strings.Trim(string(fileB), fmt.Sprintln()), fmt.Sprintln()) {
		shouldSkip, line := cleanDotenvLine(line)
		if shouldSkip {
			continue
		}

		var_split := strings.SplitN(line, "=", 2)
		if len(var_split) < 2 {
			return fileB, vars, errInvalidSegmentsDotenvLineRead(i+1, filePath)
		}

		vars[strings.Trim(strings.ToUpper(var_split[0]), " ")] = strings.TrimLeft(var_split[1], " ")
	}

	return fileB, vars, nil
}

func GetDotenv(filePath string) (types.CliConfigVars, error) {
	_, vars, err := getDotenv(filePath)
	return vars, err
}

func SetDotenv(filePath string, _vars types.CliConfigVars) error {
	vars := types.CliConfigVars{}
	maps.Copy(vars, _vars)

	fileB, _, readErr := getDotenv(filePath)
	if readErr != nil {
		return readErr
	}

	newFile := ""
	for i, line := range strings.Split(string(fileB), "\n") {
		shouldSkip, line := cleanDotenvLine(line)
		if shouldSkip {
			newFile += line + fmt.Sprintln()
			continue
		}

		var_split := strings.SplitN(line, "=", 2)
		if len(var_split) < 2 {
			return errInvalidSegmentsDotenvLineRead(i+1, filePath)
		}
		lineK := var_split[0]
		lineV := var_split[1]

		if vars[lineK] == "" {
			continue
		} else if vars[lineK] == lineV {
			newFile += line + fmt.Sprintln()
		} else {
			newFile += fmt.Sprintf("%s=%s", lineK, vars[lineK]) + fmt.Sprintln()
		}

		delete(vars, lineK)
	}
	for k, v := range vars {
		newFile += fmt.Sprintf("%s=%s", k, v) + fmt.Sprintln()
	}

	if err := os.WriteFile(filePath, []byte(newFile), 0666); err != nil {
		return err
	}

	return nil
}
