package lib

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/errs"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
)

func ValidateBranchName(branchname string) bool {
	re := regexp.MustCompile(`^[\w\-]+$`)
	if re == nil {
		panic("regular expression was not compiled!")
	}

	return re.Match([]byte(branchname))
}

func SetCurrentBranch(branchname string) error {
	configDirPath, configErr := libUtils.FindConfigDirPath()
	if configErr != nil {
		return configErr
	}

	_, cwdIsProject, cwdErr := libUtils.CwdIsProject()
	if cwdErr != nil {
		return cwdErr
	}

	if db_int.DbExists(branchname) {
		globals.CliConfig.State.CurrentBranch = branchname
		WriteConfig(globals.CliConfig, configDirPath, cwdIsProject)
	} else {
		return fmt.Errorf("provided branchname doesn't exist")
	}

	return nil
}

func NewBranchFromCurrent(dbname string, onlySchema bool) error {
	if db_int.DbExists(dbname) {
		return errs.ErrDbAlreadyExists
	}

	configFilePath, _ := libUtils.FindConfigFilePath()
	dumpFilePath := path.Join(
		libUtils.GetDriverDumpDir(configFilePath, globals.CurrentConnConfig.Driver),
		libUtils.GetDumpFileName(globals.CliConfig.State.CurrentBranch),
	)

	if err := TmpSwitchDB(globals.CliConfig.State.CurrentBranch, func() error {
		return db_int.DumpDb(dumpFilePath, globals.CliConfig.State.CurrentBranch, onlySchema)
	}); err != nil {
		return err
	}

	if err := db_int.RestoreDb(dbname, dumpFilePath, true); err != nil {
		return err
	}

	if err := os.RemoveAll(dumpFilePath); err != nil {
		return err
	}
	return nil
}
