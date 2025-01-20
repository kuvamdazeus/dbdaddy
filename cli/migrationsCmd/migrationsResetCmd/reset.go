package migrationsResetCmd

import (
	"fmt"
	"os"
	"path"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/migrationsLib"
	"github.com/fossmedaddy/dbdaddy/middlewares"
	"github.com/spf13/cobra"
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:   "reset",
	Short: "sets the currently active migration as the latest, subsequent migrations are DELETED (use with caution).",
	Run:   cmdRunFn,
}

func run(cmd *cobra.Command, args []string) {
	remoteFlag, remoteFlagErr := cmd.Flags().GetBool("remote")
	if remoteFlagErr != nil {
		fmt.Println("WARNING:", remoteFlagErr)
	}
	remoteNameFlag, remoteNameFlagErr := cmd.Flags().GetString("remote-name")
	if remoteNameFlagErr != nil {
		fmt.Println("WARNING:", remoteNameFlagErr)
	}
	usingRemoteConnConfig := remoteFlag || remoteNameFlag != ""

	connConfig, connConfigErr := lib.GetSuitableConnConfig(&globals.CliConfig, remoteFlag, remoteNameFlag)
	if connConfigErr != nil {
		fmt.Println(connConfigErr)
		return
	}
	if err := lib.TmpSwitchConn(connConfig, func() error {
		currentState, err := db_int.GetDbSchema()
		if err != nil {
			return err
		}

		migStat, err := migrationsLib.Status(currentState, usingRemoteConnConfig)
		if err != nil {
			return err
		}

		if migStat.ActiveMigration == nil {
			return fmt.Errorf("no active migration found, please make sure your migrations are up to date")
		} else if migStat.ActiveMigration.Up == nil {
			cmd.Println("already on latest migration, nothing to reset.")
			return nil
		}

		if err := os.Remove(path.Join(migStat.ActiveMigration.DirPath, constants.MigDirUpSqlFile)); err != nil {
			return err
		}

		rmMigs := []string{}
		migI := migStat.ActiveMigration.Up
		for {
			if err := os.RemoveAll(migI.DirPath); err != nil {
				return err
			}

			rmMigs = append(rmMigs, migI.DirPath)

			if migI.Up == nil {
				break
			}
			migI = migI.Up
		}

		cmd.Println("performed reset, removed migrations:")
		cmd.Println(rmMigs, fmt.Sprintln())
		cmd.Println(fmt.Sprintf("latest migration: %s", migStat.ActiveMigration.DirPath))
		return nil
	}); err != nil {
		cmd.PrintErrln(err)
		return
	}
}

func Init() *cobra.Command {
	// flags here

	return cmd
}
