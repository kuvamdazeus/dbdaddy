package migrationsCmd

import (
	"github.com/fossmedaddy/dbdaddy/cli/migrationsCmd/migrationsDownCmd"
	"github.com/fossmedaddy/dbdaddy/cli/migrationsCmd/migrationsGenCmd"
	"github.com/fossmedaddy/dbdaddy/cli/migrationsCmd/migrationsResetCmd"
	"github.com/fossmedaddy/dbdaddy/cli/migrationsCmd/migrationsStatusCmd"
	"github.com/fossmedaddy/dbdaddy/cli/migrationsCmd/migrationsUpCmd"
	"github.com/fossmedaddy/dbdaddy/lib/cliUtils"

	"github.com/spf13/cobra"
)

var (
	remoteFlag     bool
	remoteNameFlag string
)

var cmd = &cobra.Command{
	Use:   "migrations",
	Short: "manage migrations, see 'migrations --help' for more info.",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func Init() *cobra.Command {
	// flags
	cliUtils.AddRemoteFlags(cmd, &remoteFlag, &remoteNameFlag)

	// commands
	cmd.AddCommand(migrationsGenCmd.Init())
	cmd.AddCommand(migrationsStatusCmd.Init())
	cmd.AddCommand(migrationsUpCmd.Init())
	cmd.AddCommand(migrationsDownCmd.Init())
	cmd.AddCommand(migrationsResetCmd.Init())

	return cmd
}
