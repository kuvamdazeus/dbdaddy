package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fossmedaddy/dbdaddy/cli/checkoutCmd"
	"github.com/fossmedaddy/dbdaddy/cli/cloneCmd"
	"github.com/fossmedaddy/dbdaddy/cli/configCmd"
	"github.com/fossmedaddy/dbdaddy/cli/deleteCmd"
	"github.com/fossmedaddy/dbdaddy/cli/dumpCmd"
	"github.com/fossmedaddy/dbdaddy/cli/dumpMeCmd"
	"github.com/fossmedaddy/dbdaddy/cli/execCmd"
	"github.com/fossmedaddy/dbdaddy/cli/initCmd"
	"github.com/fossmedaddy/dbdaddy/cli/inspectMeCmd"
	"github.com/fossmedaddy/dbdaddy/cli/listCmd"
	"github.com/fossmedaddy/dbdaddy/cli/migrationsCmd"
	"github.com/fossmedaddy/dbdaddy/cli/remoteCmd"
	"github.com/fossmedaddy/dbdaddy/cli/restoreCmd"
	"github.com/fossmedaddy/dbdaddy/cli/schemaCmd"
	"github.com/fossmedaddy/dbdaddy/cli/statusCmd"
	"github.com/fossmedaddy/dbdaddy/cli/uriCmd"
	"github.com/fossmedaddy/dbdaddy/cli/versionCmd"
	"github.com/fossmedaddy/dbdaddy/db"
	"github.com/fossmedaddy/dbdaddy/devutils"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/manifoldco/promptui"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "dbdaddy",
	Short:            "DBDaddy is a database management tool.",
	Run:              rootCmdRun,
	PersistentPreRun: rootPreRun,
}

func rootPreRun(cmd *cobra.Command, args []string) {
	fullCmdPath := strings.Split(cmd.CommandPath(), " ")
	cmdPath := strings.Join(fullCmdPath[1:], " ")

	// DANGEROUS, FUCKED WHEN COMMAND CHANGE
	if strings.HasPrefix(cmdPath, "help") {
		return
	}
	if strings.HasSuffix(cmdPath, "--help") {
		return
	}
	if strings.HasPrefix(cmdPath, "config") {
		return
	}
	if strings.HasPrefix(cmdPath, "clone") {
		return
	}
	if strings.HasPrefix(cmdPath, "init") {
		return
	}
	if strings.HasPrefix(cmdPath, "version") {
		return
	}
	if strings.HasPrefix(cmdPath, "uri") {
		return
	}

	if lib.IsFirstTimeUser() {
		cliConfig := lib.InitCliConfig()

		cmd.Println(fmt.Sprintf("Daddy's home baby. (version: %s)", globals.Version))
		cmd.Println("I'll create a global config for ya, let me know your database uri here...")

		dbUrlPrompt := promptui.Prompt{
			Label: "(leave blank to go with a placeholder uri)",
		}

		promptConnUri, err := dbUrlPrompt.Run()
		if err != nil {
			cmd.PrintErrln("Cancelling initialization...")
			os.Exit(1)
		}
		promptConnUri = strings.Trim(promptConnUri, " ")

		if len(promptConnUri) > 0 {
			if cc, err := libUtils.GetConnConfigFromUri(promptConnUri); err != nil {
				cmd.PrintErrln("error occured while parsing connection uri")
				cmd.PrintErrln(err)
				os.Exit(1)
			} else {
				if _, err := db.ConnectDb(cc); err != nil {
					cmd.PrintErrln("can't connect to your database, please check uri...")
					cmd.PrintErrln(err)
					os.Exit(1)
				}

				cliConfig.MainConnConfig = cc
			}

		}

		configFilePath := libUtils.GetGlobalConfigPath()
		configDirPath := path.Dir(configFilePath)
		libUtils.EnsureDirExists(configDirPath)

		globals.CliConfig = cliConfig
		if err := lib.WriteConfig(cliConfig, configDirPath, false); err != nil {
			cmd.PrintErrln("error occured while initializing config file!")
			cmd.PrintErrln(err)
			os.Exit(1)
		}

		if err := libUtils.OpenFileInEditor(configFilePath); err != nil {
			cmd.Println(
				fmt.Sprintf(
					"WARNING: error occured while opening config file in a system editor. %s",
					err.Error(),
				),
			)
		}
		cmd.Println(fmt.Sprintf("config file saved at '%s'", configFilePath))
		cmd.Println(fmt.Sprintln())
	}

	if err := lib.EnsureSupportedDbDriver(); err != nil {
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	if _, err := db.ConnectDb(globals.CliConfig.MainConnConfig); err != nil {
		cmd.PrintErrln("can't connect to your database, please check uri...")
		cmd.PrintErrln(err)
		os.Exit(1)
	}

}

func rootCmdRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func main() {
	if !lib.IsFirstTimeUser() {
		configDirPath, configErr := libUtils.FindConfigDirPath()
		if configErr != nil {
			rootCmd.PrintErrln("unexpected error occured!", configErr)
			os.Exit(1)
		}

		if _, err := lib.ReadConfig(configDirPath, true); err != nil {
			rootCmd.PrintErrln("error occured while reading config!", err)
			os.Exit(1)
		}

		fmt.Println("read from configDirPath", configDirPath)
		devutils.PrettyPrint(globals.CliConfig)
		fmt.Println()
	}

	rootCmd.AddCommand(versionCmd.Init())
	rootCmd.AddCommand(checkoutCmd.Init())
	rootCmd.AddCommand(statusCmd.Init())
	rootCmd.AddCommand(deleteCmd.Init())
	rootCmd.AddCommand(configCmd.Init())
	rootCmd.AddCommand(dumpMeCmd.Init())
	rootCmd.AddCommand(dumpCmd.Init())
	rootCmd.AddCommand(listCmd.Init())
	rootCmd.AddCommand(restoreCmd.Init())
	rootCmd.AddCommand(inspectMeCmd.Init())
	// rootCmd.AddCommand(studioCmd.Init())
	rootCmd.AddCommand(execCmd.Init())
	rootCmd.AddCommand(migrationsCmd.Init())
	rootCmd.AddCommand(initCmd.Init())
	rootCmd.AddCommand(remoteCmd.Init())
	rootCmd.AddCommand(cloneCmd.Init())
	rootCmd.AddCommand(uriCmd.Init())
	rootCmd.AddCommand(schemaCmd.Init())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
