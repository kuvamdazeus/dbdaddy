package deleteCmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/middlewares"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	silentFlag bool
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:   "delete <branchname>",
	Short: "Deletes a database branch, this action can not be undone",
	Args:  cobra.MaximumNArgs(1),
	Run:   cmdRunFn,
}

func run(cmd *cobra.Command, args []string) {
	var delDbName string
	if len(args) > 0 {
		delDbName = args[0]

		if !db_int.DbExists(delDbName) {
			cmd.PrintErrf("Database branch '%s' does not exist\n", delDbName)
			return
		}
	} else {
		dbs, err := db_int.GetExistingDbs(false)
		if err != nil {
			cmd.Println("unexpected error occured", err)
			return
		}

		dbPrompt := promptui.Select{
			Label: "Choose database to delete",
			Items: dbs,
			Searcher: func(input string, index int) bool {
				return strings.Contains(dbs[index], strings.ToLower(strings.Trim(input, " ")))
			},
			StartInSearchMode: true,
		}

		_, result, pErr := dbPrompt.Run()
		if pErr != nil {
			cmd.PrintErrln("Error occured while choosing deletion database", pErr)
			return
		}

		delDbName = result
	}

	if !silentFlag {
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Deleting database: '%s' continue", delDbName),
			IsConfirm: true,
		}

		_, err := prompt.Run()
		if err != nil {
			cmd.Println("Cancelling deletion.")
			return
		}
	}

	err := db_int.DeleteDb(delDbName)
	if err != nil {
		cmd.PrintErrln("Error deleting database\n" + err.Error())
		return
	}

	// config dir directories marking
	configFilePath, _ := libUtils.FindConfigFilePath()
	configDirPath := path.Dir(configFilePath)
	// dumps marking
	driverDumpDir := path.Join(
		libUtils.GetDriverDumpDir(configFilePath, globals.CurrentConnConfig.Driver),
	)
	if dirEntries, err := os.ReadDir(driverDumpDir); err != nil {
		cmd.PrintErrln("WARNING: database dump cleanup not possible due to unexpected error")
	} else {
		for _, dirEntry := range dirEntries {
			if strings.HasSuffix(dirEntry.Name(), delDbName) {
				dumpFilePath := path.Join(driverDumpDir, dirEntry.Name())
				os.Rename(dumpFilePath, dumpFilePath+constants.SoftDeleteSuffix)
			}
		}
	}
	// migrations dir marking
	migDirName := libUtils.GetLocalMigrationsDir(path.Dir(configFilePath), delDbName)
	os.Rename(migDirName, migDirName+constants.SoftDeleteSuffix)

	// remote origin cleanup
	delete(globals.CliConfig.Origins, delDbName)
	if err := lib.WriteConfig(globals.CliConfig, configDirPath, false); err != nil {
		cmd.PrintErrln(err)
		return
	}

	cmd.Printf("Successfully deleted db '%s'\n", delDbName)

	if globals.CliConfig.State.CurrentBranch == delDbName {
		existingDbs, dbErr := db_int.GetExistingDbs(false)
		if dbErr != nil {
			cmd.PrintErrln("unexpected error occured!", dbErr)
			os.Exit(1)
		}

		newBranchName := existingDbs[0]
		globals.CliConfig.State.CurrentBranch = newBranchName
		if err := lib.WriteConfig(globals.CliConfig, configDirPath, false); err != nil {
			cmd.PrintErrln("unexpected error occured!", err)
			os.Exit(1)
		}

		cmd.Printf("Switched to database branch '%s'\n", newBranchName)
	}
}

func Init() *cobra.Command {
	cmd.Flags().BoolVar(&silentFlag, "silent", false, "silences confirmation for deleting branch")

	return cmd
}
