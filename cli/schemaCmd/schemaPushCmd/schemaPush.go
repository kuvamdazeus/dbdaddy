package schemaPushCmd

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/cliUtils"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/lib/migrationsLib"
	"github.com/fossmedaddy/dbdaddy/middlewares"
	"github.com/fossmedaddy/dbdaddy/sqlwriter"
	"github.com/fossmedaddy/dbdaddy/types"
	"github.com/spf13/cobra"
)

var (
	dryRunFlag bool
	forceFlag  bool
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:   "push",
	Short: "push & override database with changes done in schema folder",
	Run:   cmdRunFn,
}

func run(cmd *cobra.Command, args []string) {
	cwd, cwdIsProject, err := libUtils.CwdIsProject()
	if err != nil {
		cmd.PrintErrln("unexpected error occured while reading CWD")
		cmd.PrintErrln(err)
		return
	}
	if !cwdIsProject {
		cmd.PrintErrln("CWD is not a project! initialize one with 'init' command")
		return
	}

	if err := cliUtils.TmpSwitchSuitableConn(cmd, func(connConfig types.ConnConfig, usingRemoteConnConfig bool) error {
		var wg sync.WaitGroup

		stmts := []string{}
		schemaDirPath := path.Join(cwd, constants.SchemaDirName)
		filepath.WalkDir(schemaDirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || !strings.HasSuffix(d.Name(), ".sql") {
				return nil
			}

			if fileB, err := os.ReadFile(path); err != nil {
				return err
			} else {
				fileSqlStmts := libUtils.GetSQLStmts(string(fileB) + fmt.Sprintln())
				stmts = slices.Concat(stmts, fileSqlStmts)
			}

			return nil
		})

		disableConstSql, disableConstErr := sqlwriter.GetDisableConstSQL()
		if disableConstErr != nil {
			return disableConstErr
		}
		enableConstSql, enableConstErr := sqlwriter.GetEnableConstSQL()
		if enableConstErr != nil {
			return enableConstErr
		}
		disableConstSql += fmt.Sprintln()
		enableConstSql += fmt.Sprintln()

		stmts = slices.Concat([]string{disableConstSql}, stmts, []string{enableConstSql})

		var userDefinedSchema *types.DbSchema
		if err := lib.TmpSwitchToShadowDB(func() error {
			if err := db_int.ExecuteStatementsTx(stmts); err != nil {
				cmd.Println("WARNING: error occured in schema definition")
				return err
			}

			if schema, err := db_int.GetDbSchema(); err != nil {
				cmd.Println("WARNING: error occured while fetching schema for shadow database")
				return err
			} else {
				userDefinedSchema = schema
			}

			return nil
		}); err != nil {
			return err
		}
		userDefinedSchema.DbName = connConfig.Database

		dbSchema, err := db_int.GetDbSchema()
		if err != nil {
			return err
		}

		latestMig, isMigInit, migErr := migrationsLib.GetLatestMigrationOrInit(dbSchema, "", usingRemoteConnConfig)
		if migErr != nil {
			return migErr
		}
		if !latestMig.IsActive {
			if isMigInit {
				cmd.Println("WARNING: it is advised to generate migrations for your database in order to safely track and revert all the changes you make from schema changes")
			} else {
				cmd.Println("WARNING: you are not at the latest migration, this is dangerous for your database, it is advised to run this operation from latest migration")
			}

			if !forceFlag {
				return fmt.Errorf("invalid operation not allowed")
			}
		}

		var (
			upSqlScript, downSqlScript string
			upChanges                  []types.MigAction
		)

		var (
			upSqlErr   error
			downSqlErr error
		)
		(func() {
			wg.Add(2)
			defer wg.Wait()

			go (func() {
				defer wg.Done()
				upChanges = migrationsLib.DiffDBSchema(userDefinedSchema, dbSchema)
				upSqlScript, upSqlErr = migrationsLib.GetSQLFromDiffChanges(upChanges)
			})()

			go (func() {
				defer wg.Done()
				downChanges := migrationsLib.DiffDBSchema(dbSchema, userDefinedSchema)
				downSqlScript, downSqlErr = migrationsLib.GetSQLFromDiffChanges(downChanges)
			})()
		})()
		if upSqlErr != nil || downSqlErr != nil {
			return fmt.Errorf(fmt.Sprintln(upSqlErr) + fmt.Sprintln(downSqlErr))
		}

		if len(upChanges) == 0 {
			cmd.Println("your schema folder is up-to-date with database schema")
			return nil
		}

		downSqlScript = disableConstSql + downSqlScript + enableConstSql

		if dryRunFlag {
			cmd.Println(constants.DownSqlScriptComment)
			cmd.Println(downSqlScript)

			cmd.Println(constants.UpSqlScriptComment)
			cmd.Println(upSqlScript)

			return nil
		}

		upStmts := []string{}
		if sql, err := sqlwriter.GetDisableConstSQL(); err != nil {
			return err
		} else {
			upStmts = append(upStmts, sql)
		}
		upStmts = slices.Concat(upStmts, libUtils.GetSQLStmts(upSqlScript))

		if err := db_int.ExecuteStatementsTx(upStmts); err != nil {
			return err
		}

		opts := migrationsLib.GenMigOpts{
			CurrentState:    userDefinedSchema,
			LatestMigration: latestMig,
			UpSql:           upSqlScript,
			DownSql:         downSqlScript,
			InfoStr:         migrationsLib.GetInfoTextFromDiff(upChanges),
			UseRemoteDir:    usingRemoteConnConfig,
		}
		_, migGenErr := migrationsLib.GenerateMigration(opts)
		if migGenErr != nil {
			return migGenErr
		}

		cmd.Println("pushed schema successfully.")
		return nil
	}); err != nil {
		cmd.PrintErrln(err)
		return
	}
}

func Init() *cobra.Command {
	// flags here
	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "prints out sql queries for diffed changes, no migrations are generated or run.")
	cmd.Flags().BoolVar(&forceFlag, "force", false, "bypass checks that prevent you from corrupting your database history")

	return cmd
}
