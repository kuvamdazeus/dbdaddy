package restoreCmd

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/middlewares"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	useGlobalContext bool
	userDumpFilePath string
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:   "restore",
	Short: "restores the current database branch with a given database dump file, does so by OVER-WRITING THE EXISTING CONTENTS.",
	Run:   cmdRunFn,
}

func run(cmd *cobra.Command, args []string) {
	var dumpFilePath string
	if len(userDumpFilePath) == 0 {
		var configFilePath string
		if useGlobalContext {
			configFilePath = libUtils.GetGlobalConfigPath()
		} else {
			configFilePath, _ = libUtils.FindConfigFilePath()
		}
		dbGroupedDumpFiles, err := lib.GetDbGroupedDumpFiles(configFilePath)
		if err != nil {
			cmd.PrintErrln("Error occured while fetching available dump files!\n", err)
			return
		}
		allDumps := []string{}
		for _, dumpFiles := range dbGroupedDumpFiles {
			allDumps = slices.Concat(allDumps, dumpFiles)
		}

		prompt := promptui.Select{
			Label:             "Available dumps to restore from",
			Items:             allDumps,
			StartInSearchMode: true,
			Searcher: func(input string, index int) bool {
				return strings.Contains(allDumps[index], input)
			},
		}

		_, promptDumpFilePath, err := prompt.Run()
		if err != nil {
			cmd.PrintErrln(err)
			return
		}

		dumpFilePath = promptDumpFilePath
	} else {
		path, err := filepath.Abs(userDumpFilePath)
		if err != nil {
			cmd.PrintErrln("Unexpected error occured!\n" + err.Error())
			return
		}

		dumpFilePath = path
	}

	cmd.Printf("Attempting restore on db: '%s' using backup dump file: '%s'...\n", globals.CliConfig.State.CurrentBranch, dumpFilePath)
	if err := lib.TmpSwitchDB(globals.CliConfig.State.CurrentBranch, func() error {
		return db_int.RestoreDb(globals.CliConfig.State.CurrentBranch, dumpFilePath, true)
	}); err != nil {
		cmd.PrintErrln("error occured while restoring db\n", err)
		return
	}

	cmd.Println("Success!")
}

func Init() *cobra.Command {
	cmd.Flags().StringVar(&userDumpFilePath, "file", "", "provide a dump file path to restore the current db branch from")
	cmd.Flags().BoolVarP(&useGlobalContext, "global", "g", false, "use global context to list dumps from.")

	return cmd
}
