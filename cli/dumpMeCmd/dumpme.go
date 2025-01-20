package dumpMeCmd

import (
	"errors"
	"os"
	"path"

	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/errs"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/middlewares"

	"github.com/spf13/cobra"
)

var (
	useGlobalFile = false
	// outputPathArg = ""
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:   "dumpme",
	Short: "Takes a dump of the current database branch, this dump file can be later used to restore the data",
	Run:   cmdRunFn,
}

func run(cmd *cobra.Command, args []string) {
	configFilePath, _ := libUtils.FindConfigFilePath()
	if useGlobalFile {
		configFilePath = libUtils.GetGlobalConfigPath()
	}

	connConfig := globals.CurrentConnConfig
	connConfig.Database = globals.CliConfig.State.CurrentBranch

	outputFilePath := path.Join(
		libUtils.GetDriverDumpDir(configFilePath, connConfig.Driver),
		libUtils.GetDumpFileName(connConfig.Database),
	)

	if err := lib.TmpSwitchConn(connConfig, func() error {
		dumpErr := db_int.DumpDb(outputFilePath, connConfig.Database, false)
		if dumpErr != nil {
			if errors.Is(dumpErr, errs.ErrPgDumpCmdNotFound) {
				cmd.Println("Hey! we noticed you don't have 'pg_dump', then you also probably won't have 'pg_restore', we use these tools internally to perform dumps & restores... please install these tools in your OS before proceeding.")
			}
		}

		return dumpErr
	}); err != nil {
		cmd.PrintErrln("unexpected error occured!")
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	cmd.Println("\nDumped successfully! output file:", outputFilePath)
}

func Init() *cobra.Command {
	cmd.Flags().BoolVarP(&useGlobalFile, "global", "g", false, "explicitly use global config file creds to connect to db")
	// cmd.Flags().StringVarP(&outputPathArg, "output", "o", "", "define output dump file path explicitly, by default i'll store it with me at your nearest '.dbdaddy' directory")

	return cmd
}
