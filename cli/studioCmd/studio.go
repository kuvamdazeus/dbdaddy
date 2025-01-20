package studioCmd

import (
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	libServer "github.com/fossmedaddy/dbdaddy/lib/server"
	"github.com/fossmedaddy/dbdaddy/middlewares"

	"github.com/spf13/cobra"
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:     "studio",
	Aliases: []string{"soymedaddy"},
	Short:   "opens up a UI to manage your data & run queries",
	Run:     cmdRunFn,
}

func run(cmd *cobra.Command, args []string) {
	lib.TmpSwitchDB(globals.CliConfig.State.CurrentBranch, func() error {
		cmd.Println("Starting webserver at: http://127.0.0.1:42069")
		cmd.Println("Once, server is up, head towards: https://dbdaddy.hackerrizz.com/studio")
		if err := libServer.StartServer(); err != nil {
			cmd.Println("Could not start server!", err.Error())
			return nil
		}

		return nil
	})
}

func Init() *cobra.Command {
	// add flags

	return cmd
}
