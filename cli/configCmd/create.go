package configCmd

import (
	"fmt"
	"os"
	"path"

	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"

	"github.com/spf13/cobra"
)

var createInGlobalNamespace = false
var overrideExisting = false

var createCmd = &cobra.Command{
	Use:   "create",
	Short: fmt.Sprintf("Create a new config file in the current working dir. Use '-g or --global' to create config file at %s", libUtils.GetGlobalConfigPath()),
	Run:   createCmdRun,
}

func createCmdRun(cmd *cobra.Command, args []string) {
	configWritePath := libUtils.GetLocalConfigPath()
	if createInGlobalNamespace {
		configWritePath = libUtils.GetGlobalConfigPath()
	}

	if !overrideExisting {
		if _, err := os.Stat(configWritePath); err == nil {
			cmd.PrintErrln(fmt.Sprintf("config file at '%s' already exists!", configWritePath))
			return
		}
	}

	configWriteDirPath, _ := path.Split(configWritePath)
	if _, err := libUtils.EnsureDirExists(configWriteDirPath); err != nil {
		panic("Unexpected error occured!\n" + err.Error())
	}

	cliConfig := lib.InitCliConfig()
	if err := lib.WriteConfig(cliConfig, configWriteDirPath, true); err != nil {
		cmd.PrintErrln("Error occured while writing config file!\n" + err.Error())
		return
	}

	libUtils.OpenFileInEditor(configWritePath)
}

func InitCreateCmd() *cobra.Command {
	createCmd.Flags().BoolVarP(&createInGlobalNamespace, "global", "g", false, fmt.Sprintf("write config file at '%s' for global use, writes to '%s' by default.", libUtils.GetGlobalConfigPath(), libUtils.GetLocalConfigPath()))
	createCmd.Flags().BoolVarP(&overrideExisting, "force", "f", false, "use force. override any existing config files present.")

	return createCmd
}
