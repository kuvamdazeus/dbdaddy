package remoteLsCmd

import (
	"fmt"

	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "ls",
	Short: "list all available remote origins",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	if len(globals.CliConfig.Origins) == 0 {
		cmd.Println("no available origins.")
		return
	}

	cmd.Println("listing available origins for databases")

	i := 1
	for originKey, originConnConfig := range globals.CliConfig.Origins {
		cmd.Println(fmt.Sprintf("%d. %s <-> %s:%s/%s", i, originKey, originConnConfig.Host, originConnConfig.Port, originConnConfig.Database))
		i++
	}
}

func Init() *cobra.Command {
	// flags here

	return cmd
}
