package schemaPullCmd

import (
	"fmt"
	"path"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/cliUtils"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/middlewares"
	"github.com/fossmedaddy/dbdaddy/types"
	"github.com/spf13/cobra"
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:   "pull",
	Short: "DELETES the current schema directory and re-creates it, populating it with latest schema from database",
	Run:   cmdRunFn,
}

func run(cmd *cobra.Command, args []string) {
	cwd, cwdIsProject, cwdErr := libUtils.CwdIsProject()
	if cwdErr != nil {
		cmd.PrintErrln("unexpected error occured")
		cmd.PrintErrln(cwdErr)
		return
	}

	if !cwdIsProject {
		cmd.Println("CWD is not a project, please create a project with 'init' command or cd into one.")
		return
	}

	if err := lib.EnsureProjectDirsExist(); err != nil {
		cmd.PrintErrln("unexpected error occured")
		cmd.PrintErrln(err)
		return
	}

	schemaDirPath := path.Join(cwd, constants.SchemaDirName)

	if err := cliUtils.TmpSwitchSuitableConn(cmd, func(connConfig types.ConnConfig, usingRemoteConnConfig bool) error {
		if err := lib.LoadDbSchemaIntoSchemaDir(schemaDirPath); err != nil {
			return err
		}

		return nil
	}); err != nil {
		cmd.PrintErrln("unexpected error occured")
		cmd.PrintErrln(err)
		return
	}

	cmd.Println(fmt.Sprintf("database schema loaded into '%s' directory successfully", schemaDirPath))
}

func Init() *cobra.Command {
	// flags here

	return cmd
}
