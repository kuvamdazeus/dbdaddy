package remoteAddCmd

import (
	"fmt"
	"slices"

	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/middlewares"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

const cmdManual = `
Name of the origin is same as the database name the origin is being set up of for.
`

var (
	forceFlag bool
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:     "add",
	Short:   "add a remote origin by a providing a name and a db connection uri for the current branch",
	Example: "add postgresql://user:pwd@localhost:5432/dbname",
	Run:     cmdRunFn,
	Args:    cobra.ExactArgs(1),
}

func run(cmd *cobra.Command, args []string) {
	configDirPath, _ := libUtils.FindConfigDirPath()

	connUri := args[0]
	connConfig, uriErr := libUtils.GetConnConfigFromUri(connUri)
	if uriErr != nil {
		cmd.Println(fmt.Sprintf("error occured while parsing database url: %s", uriErr.Error()))
		return
	}
	originName := connConfig.Database

	if slices.Contains(maps.Keys(globals.CliConfig.Origins), originName) && !forceFlag {
		cmd.Println(
			fmt.Sprintf(
				"remote origin for '%s' already exists in the config, use force flag to override",
				globals.CliConfig.State.CurrentBranch,
			),
		)
		return
	}

	if err := lib.TmpSwitchConn(connConfig, func() error {
		return nil
	}); err != nil {
		cmd.PrintErrln(err)
		return
	}

	globals.CliConfig.Origins[originName] = connConfig
	if err := lib.WriteConfig(globals.CliConfig, configDirPath, true); err != nil {
		cmd.PrintErrln("error occured while saving config:", err)
		return
	}

	cmd.Println(fmt.Sprintf("remote origin for '%s' successfully set.", globals.CliConfig.State.CurrentBranch))
}

func Init() *cobra.Command {
	// flags here
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "will override existing configuration for provided origin")

	return cmd
}
