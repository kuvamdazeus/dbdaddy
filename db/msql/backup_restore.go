package msql

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fossmedaddy/dbdaddy/globals"
)

func DumpDb(outputFilePath string, onlySchema bool) error {
	args := []string{
		// fmt.Sprintf("--user=%s", v.GetString(constants.DbConfigUserKey)),
		// fmt.Sprintf("--password=%s", v.GetString(constants.DbConfigPassKey)),
		// fmt.Sprintf("--host=%s", v.GetString(constants.DbConfigHostKey)),
		// fmt.Sprintf("--port=%s", v.GetString(constants.DbConfigPortKey)),
		// fmt.Sprintf("--result-file=%s", outputFilePath),
		"--skip-comments",
		"--no-create-db",
		"--no-autocommit",
		"--databases",
		globals.CliConfig.State.CurrentBranch,
	}
	if onlySchema {
		args = append(args, "--no-data")
	}
	dumpCmd := exec.Command(
		"mysqldump", args...,
	)

	pipe, err := dumpCmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := dumpCmd.Start(); err != nil {
		return err
	}

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			return err
		}
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), "use ") {
			continue
		}

		if _, err := outputFile.WriteString(line + fmt.Sprintln()); err != nil {
			return err
		}
	}

	dumpCmd.Wait()
	return nil
}

func RestoreDb(dbname string, dumpFilePath string, override bool) error {
	if DbExists(dbname) {
		if override {
			if err := DeleteDb(dbname); err != nil {
				return err
			}

			if err := CreateDb(dbname); err != nil {
				return err
			}
		}
	} else {
		if err := CreateDb(dbname); err != nil {
			return err
		}
	}

	dumpFile, fileErr := os.Open(dumpFilePath)
	if fileErr != nil {
		return fileErr
	}

	restoreCmd := exec.Command(
		"mysql",
		// fmt.Sprintf("--user=%s", v.GetString(constants.DbConfigUserKey)),
		// fmt.Sprintf("--password=%s", v.GetString(constants.DbConfigPassKey)),
		// fmt.Sprintf("--host=%s", v.GetString(constants.DbConfigHostKey)),
		// fmt.Sprintf("--port=%s", v.GetString(constants.DbConfigPortKey)),
		// fmt.Sprintf("--database=%s", dbname),
	)

	restoreCmd.Stdin = dumpFile
	restoreCmd.Stderr = os.Stderr

	if err := restoreCmd.Run(); err != nil {
		return err
	}

	return nil
}
