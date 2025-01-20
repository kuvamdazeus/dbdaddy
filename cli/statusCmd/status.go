package statusCmd

import (
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/middlewares"

	"github.com/spf13/cobra"
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of the current database branch, also pings the database.",
	Run:   cmdRunFn,
}

func run(cmd *cobra.Command, args []string) {
	cmd.Println("On branch:", globals.CliConfig.State.CurrentBranch)
}

func Init() *cobra.Command {
	// bind flags or something here
	// ...

	return cmd
}
