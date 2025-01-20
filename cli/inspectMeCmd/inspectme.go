package inspectMeCmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/lib/cliUtils"
	"github.com/fossmedaddy/dbdaddy/middlewares"
	"github.com/fossmedaddy/dbdaddy/sqlwriter"
	"github.com/fossmedaddy/dbdaddy/types"
	"golang.org/x/exp/maps"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	showAll        bool
	remoteFlag     bool
	remoteNameFlag string
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:     "inspect",
	Aliases: []string{"inspectme"},
	Short:   "prints the schema of a selected table in current database",
	Run:     cmdRunFn,
	Args:    cobra.MaximumNArgs(100),
}

func getColName(name string, pk bool) string {
	if pk {
		return fmt.Sprintf("%s (Primary Key)", name)
	}

	return name
}

func run(cmd *cobra.Command, args []string) {
	err := cliUtils.TmpSwitchSuitableConn(cmd, func(connConfig types.ConnConfig, usingRemoteConnConfig bool) error {
		selectedTables := []string{}

		dbSchema, err := db_int.GetDbSchema()
		if err != nil {
			return fmt.Errorf("unexpected error occured while fetching tables from database '%s'\n%s", connConfig.Database, err.Error())
		}

		dbStrTables := maps.Keys(dbSchema.Tables)
		slices.Sort(dbStrTables)

		notFoundTableWarnings := false
		if showAll {
			selectedTables = dbStrTables
		} else if len(args) > 0 {
			for _, arg := range args {
				var searchTableId string
				if strings.Contains(".", arg) {
					searchTableId = arg
				} else {
					searchTableId = fmt.Sprintf("public.%s", arg)
				}

				if dbSchema.Tables[searchTableId] != nil {
					selectedTables = append(selectedTables, searchTableId)
				} else {
					if !notFoundTableWarnings {
						notFoundTableWarnings = true
					}
					cmd.Println(fmt.Sprintf("WARNING: table '%s' was not found!", searchTableId))
				}
			}
		} else {
			prompt := promptui.Select{
				Label: "Choose table to display schema of",
				Items: dbStrTables,
				Searcher: func(input string, index int) bool {
					return strings.Contains(dbStrTables[index], strings.ToLower(strings.Trim(input, " ")))
				},
				StartInSearchMode: true,
			}

			_, result, err := prompt.Run()
			if err != nil {
				return err
			}

			selectedTables = append(selectedTables, result)
		}
		if notFoundTableWarnings {
			cmd.Println()
		}

		viewsPrintBuf := ""
		for _, tableid := range selectedTables {
			isView := false
			tableSchema := dbSchema.Tables[tableid]
			if tableSchema == nil {
				tableSchema = dbSchema.Views[tableid]
				isView = true
			}

			var getSqlFn func(*types.TableSchema) (string, error)
			if !isView {
				getSqlFn = sqlwriter.GetCreateTableSQL
			} else {
				getSqlFn = sqlwriter.GetCreateViewSQL
			}
			defSql, err := getSqlFn(tableSchema)
			if err != nil {
				return err
			}

			indSql := ""
			if !isView {
				for _, index := range tableSchema.Indexes {
					if sql, err := sqlwriter.GetCreateIndexSQL(&index); err != nil {
						cmd.Println(fmt.Sprintf("WARNING: sql for index '%s' cannot be generated! %s", index.Name, err.Error()))
					} else {
						indSql += sql
					}
				}
			}

			tableConSql := ""
			if !isView {
				for _, con := range tableSchema.Constraints {
					if atConSql, err := sqlwriter.GetATCreateConstraintSQL(tableid, con); err != nil {
						return err
					} else {
						tableConSql += atConSql
					}
				}
			}

			if isView {
				viewsPrintBuf += fmt.Sprintln(fmt.Sprintf("--- VIEW: %s", tableid))
				viewsPrintBuf += fmt.Sprintln(defSql + tableConSql)
				viewsPrintBuf += fmt.Sprintln()
			} else {
				cmd.Println(fmt.Sprintf("--- TABLE: %s", tableid))
				cmd.Println(defSql + tableConSql + fmt.Sprintln() + indSql)
				cmd.Println()
			}
		}

		if len(viewsPrintBuf) > 0 {
			cmd.Print(viewsPrintBuf)
		}

		return nil
	})
	if err != nil {
		cmd.PrintErrln(err)
	}
}

func Init() *cobra.Command {
	cliUtils.AddRemoteFlags(cmd, &remoteFlag, &remoteNameFlag)
	cmd.Flags().BoolVar(&showAll, "all", false, "print schema for all tables")

	return cmd
}
