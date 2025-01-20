package cloneCmd

import (
	"fmt"
	"os"
	"path"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/db"
	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/spf13/cobra"
)

var (
	forceFlag      = false
	onlySchemaFlag = false
)

var cmd = &cobra.Command{
	Use:     "clone",
	Short:   "provide a uri to take a dump of the remote database into the currently connected instance",
	Example: "clone postgresql://user:pwd@localhost:5432/dbname",
	Run:     run,
	Args:    argFn,
}

func argFn(cmd *cobra.Command, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("invalid number of arguments provided, need max: 2, min: 1 args")
	}

	return nil
}

func run(cmd *cobra.Command, args []string) {
	configDirPath, configErr := libUtils.FindConfigDirPath()
	if configErr != nil {
		cmd.PrintErrln("please setup a project or configure a global config! no local instance specified to connect to!")
		os.Exit(1)
	}
	configFilePath := path.Join(configDirPath, constants.SelfConfigFileName)

	if _, err := lib.ReadConfig(configDirPath, true); err != nil {
		cmd.PrintErrln("unexpected error occured!", err)
		os.Exit(1)
	}

	remoteUri := args[0]
	remoteConnConfig, err := libUtils.GetConnConfigFromUri(remoteUri)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	dbname := ""
	if len(args) == 2 {
		dbname = args[1]
	} else {
		dbname = remoteConnConfig.Database
	}

	cmd.Println("started dump process...")

	dumpOutFile := path.Join(
		libUtils.GetDriverDumpDir(configFilePath, remoteConnConfig.Driver),
		libUtils.GetDumpFileName(dbname),
	)

	if err := lib.TmpSwitchConn(remoteConnConfig, func() error {
		dumpErr := db_int.DumpDb(dumpOutFile, remoteConnConfig.Database, onlySchemaFlag)
		if dumpErr != nil {
			cmd.PrintErrln("error occured while taking a dump of the remote database!")
			cmd.PrintErrln(dumpErr)
		}

		return dumpErr
	}); err != nil {
		cmd.PrintErrln("unexpected error occured!")
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	cmd.Println(fmt.Sprintf("remote database dump complete, saved at: %s", dumpOutFile))

	if err := lib.TmpSwitchConn(globals.CliConfig.MainConnConfig, func() error {
		_, connectErr := db.ConnectSelfDb(globals.CliConfig.MainConnConfig)
		if connectErr != nil {
			return connectErr
		}

		return nil
	}); err != nil {
		cmd.PrintErrln("unexpected error occured while connecting to the database!")
		cmd.PrintErrln(err)
		os.Exit(1)
	}

	if err := lib.TmpSwitchConn(globals.CliConfig.MainConnConfig, func() error {
		if db_int.DbExists(dbname) && !forceFlag {
			cmd.PrintErrln(fmt.Printf("database with name %s already exists, choose a different db name or use --force flag.", dbname))
			os.Exit(1)
		}

		return db_int.RestoreDb(dbname, dumpOutFile, true)
	}); err != nil {
		cmd.PrintErrln("unexpected error occured while restoring database!", err)
		os.Exit(1)
	}

	globals.CliConfig.Origins[remoteConnConfig.Database] = remoteConnConfig
	if err := lib.WriteConfig(globals.CliConfig, configDirPath, true); err != nil {
		cmd.PrintErrln("error occured while writing to config file")
		cmd.PrintErrln(err)
		return
	}

	cmd.Println(fmt.Sprintf("remote database '%s' cloned successfully.", dbname))
}

func Init() *cobra.Command {
	// flags here
	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "force remote database copying even if there already exists a database with same name")
	cmd.Flags().BoolVar(&onlySchemaFlag, "only-schema", false, "dump only schema, exclude data from dump")

	return cmd
}
