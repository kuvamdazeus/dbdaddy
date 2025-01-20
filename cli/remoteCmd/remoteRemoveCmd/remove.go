package remoteRemoveCmd

import (
	"fmt"
	"os"

	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:     "remove",
	Short:   "remove origin(s) by name",
	Aliases: []string{"rm"},
	Args:    cobra.MinimumNArgs(1),
	Run:     run,
}

func run(cmd *cobra.Command, args []string) {
	configDirPath, configErr := libUtils.FindConfigDirPath()
	if configErr != nil {
		cmd.PrintErrln("unexpected error occured!", configErr)
		os.Exit(1)
	}

	_, cwdIsProject, cwdErr := libUtils.CwdIsProject()
	if cwdErr != nil {
		cmd.PrintErrln("unexpected error occured!", cwdErr)
		os.Exit(1)
	}

	origins := globals.CliConfig.Origins
	for _, argOrigin := range args {
		delete(origins, argOrigin)
	}

	globals.CliConfig.Origins = origins
	if err := lib.WriteConfig(globals.CliConfig, configDirPath, cwdIsProject); err != nil {
		cmd.PrintErrln(err)
		return
	}

	cmd.Println(fmt.Sprintf("Removed origins %v successfully.", args))
}

func Init() *cobra.Command {
	// flags here

	return cmd
}
