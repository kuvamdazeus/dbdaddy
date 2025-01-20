package pg

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/fossmedaddy/dbdaddy/db"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/types"
)

var (
	sharedEnv = append(os.Environ(), "PGSSLMODE=allow")
)

func DumpDb(outputFilePath string, connConfig types.ConnConfig, onlySchema bool) error {
	connUri := db.GetPgConnUriFromConnConfig(connConfig)

	dumpAllArgs := []string{
		"--globals-only",
		"--clean",
		"--if-exists",
		"--dbname=" + connUri,
		"--file=" + outputFilePath,
	}
	dumpAllCmd := exec.Command("pg_dumpall", dumpAllArgs...)

	dumpAllCmd.Env = sharedEnv
	dumpAllCmd.Stderr = os.Stderr

	if err := dumpAllCmd.Run(); err != nil {
		return fmt.Errorf(err.Error())
	}

	dumpArgs := []string{fmt.Sprintf("--dbname=%s", connUri)}
	if onlySchema {
		dumpArgs = append(dumpArgs, "--schema-only")
	}
	dumpCmd := exec.Command("pg_dump", dumpArgs...)

	dumpCmd.Env = sharedEnv
	dumpCmd.Stderr = os.Stderr

	outputFile, fileErr := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		return fileErr
	}
	defer outputFile.Close()

	if _, err := outputFile.WriteString(fmt.Sprintln()); err != nil {
		return err
	}

	pipe, fileErr := dumpCmd.StdoutPipe()
	if fileErr != nil {
		return fileErr
	}

	if err := dumpCmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		if scanner.Err() != nil {
			return scanner.Err()
		}

		line := scanner.Text()
		if _, err := outputFile.WriteString(line + fmt.Sprintln()); err != nil {
			return err
		}
	}

	if err := dumpCmd.Wait(); err != nil {
		return err
	}

	return nil
}

// creates db & restores its content from given dump
func RestoreDb(dbname string, dumpFilePath string, override bool) error {
	connConfig := globals.CurrentConnConfig
	connConfig.Database = dbname

	if DbExists(connConfig.Database) {
		if override {
			if err := DeleteDb(connConfig.Database); err != nil {
				return err
			}

			if err := CreateDb(connConfig.Database); err != nil {
				return err
			}
		}
	} else {
		if err := CreateDb(connConfig.Database); err != nil {
			return err
		}
	}

	// run restore command
	osCmd := exec.Command(
		"psql",
		"--dbname="+db.GetPgConnUriFromConnConfig(connConfig),
		"--file="+dumpFilePath,
	)

	osCmd.Env = sharedEnv
	osCmd.Stderr = os.Stderr

	err := osCmd.Run()
	if err != nil {
		return err
	}

	return nil
}
