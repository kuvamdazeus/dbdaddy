package listCmd

import (
	"fmt"

	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib/cliUtils"
	"github.com/fossmedaddy/dbdaddy/middlewares"
	"github.com/fossmedaddy/dbdaddy/types"

	"github.com/spf13/cobra"
)

var (
	showHiddenFlag bool
	remoteFlag     bool
	remoteNameFlag string
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "Lists available database branches on the db server",
	Run:     cmdRunFn,
}

func run(cmd *cobra.Command, args []string) {
	var dbs []string
	err := cliUtils.TmpSwitchSuitableConn(cmd, func(connConfig types.ConnConfig, usingRemoteConnConfig bool) error {
		if _dbs, err := db_int.GetExistingDbs(showHiddenFlag); err != nil {
			return err
		} else {
			dbs = _dbs
			return nil
		}
	})
	if err != nil {
		cmd.PrintErrln("Unexpected error occured!\n" + err.Error())
		return
	}

	cmd.Println("Available database branches:")
	for i, db := range dbs {
		dbStr := fmt.Sprintf("%d - %s", i+1, db)
		if db == globals.CliConfig.State.CurrentBranch {
			dbStr += " (current branch)"
		}

		cmd.Println(dbStr)
	}
}

func Init() *cobra.Command {
	// add flags
	cliUtils.AddRemoteFlags(cmd, &remoteFlag, &remoteNameFlag)

	cmd.Flags().BoolVarP(&showHiddenFlag, "show-hidden", "s", false, "Show hidden databases")

	return cmd
}
