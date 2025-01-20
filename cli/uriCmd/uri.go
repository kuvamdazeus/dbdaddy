package uriCmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/types"
	"github.com/spf13/cobra"
)

var (
	useGlobalConfigFlag bool
	useShadowConfigFlag bool
)

var cmd = &cobra.Command{
	Use:     "uri",
	Short:   "replace the connection credentials in dbdaddy config with a newly provided db connection uri",
	Example: "uri postgresql://user:pwd@127.0.0.1:5432/dbname",
	Args:    cobra.ExactArgs(1),
	Run:     run,
}

func run(cmd *cobra.Command, args []string) {
	uri := args[0]
	uri = strings.Trim(uri, " ")

	connConfig := types.ConnConfig{}
	if cc, err := libUtils.GetConnConfigFromUri(uri); err != nil {
		cmd.PrintErrln(err)
		return
	} else {
		connConfig = cc
	}

	if err := lib.TmpSwitchConn(connConfig, func() error {
		return nil
	}); err != nil {
		cmd.PrintErrln(fmt.Sprintf("could not connect to '%s'", connConfig.ConnString))
		cmd.PrintErrln("please check your provided connection uri.")
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	configDirPath, err := libUtils.FindConfigDirPath()
	if err != nil {
		cmd.PrintErrln("error occured while trying to find config file in your device...")
		cmd.PrintErrln(err)
		return
	}

	if useGlobalConfigFlag {
		configDirPath = libUtils.GetGlobalConfigPath()
	}

	if useShadowConfigFlag {
		globals.CliConfig.ShadowConnConfig = &connConfig
	} else {
		globals.CliConfig.MainConnConfig = connConfig
		globals.CliConfig.State.CurrentBranch = connConfig.Database
	}

	_, isCwdProject, err := libUtils.CwdIsProject()
	if err != nil {
		cmd.PrintErrln("unexpected error occured!")
		os.Exit(1)
	}

	if err := lib.WriteConfig(globals.CliConfig, configDirPath, isCwdProject); err != nil {
		cmd.PrintErrln("unexpected error occured while writing to config file")
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	cmd.Println("db connection credentials changed successfully in config")
}

func Init() *cobra.Command {
	// flags here
	cmd.Flags().BoolVarP(&useGlobalConfigFlag, "global", "g", false, "explicitly use global config file")
	cmd.Flags().BoolVar(&useShadowConfigFlag, "shadow", false, "the provided db connection uri will be used for connecting to your dedicated shadow database")

	return cmd
}
