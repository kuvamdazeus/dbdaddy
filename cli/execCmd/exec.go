package execCmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/cliUtils"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/middlewares"
	"github.com/fossmedaddy/dbdaddy/types"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	noTx           bool
	outFileFlag    string
	queryFlag      string
	remoteFlag     bool
	remoteNameFlag string
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:     "exec",
	Short:   "Run SQL queries directly from CLI and get query outputs in CSV format",
	Example: "usage examples: 'exec my_queries/create_order_table.sql', 'exec -q \"select * from public.users\" -o ./my_exports/users.csv'",
	Run:     cmdRunFn,
	Args:    cobra.MaximumNArgs(1),
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
	return string(s)
}

func runFile(cmd *cobra.Command, filename string) error {
	cwd, cwdIsProject, err := libUtils.CwdIsProject()
	if err != nil {
		return err
	}

	sqlFileName := filename
	if !strings.HasSuffix(sqlFileName, ".sql") {
		sqlFileName += ".sql"
	}

	queryWd := cwd
	if cwdIsProject {
		queryWd = path.Join(queryWd, constants.ScriptsDirName)
	}

	runSqlPath := path.Join(queryWd, sqlFileName)
	if !libUtils.Exists(runSqlPath) {
		cmd.PrintErrln(fmt.Sprintf("file '%s' does not exist", runSqlPath))
		if cwdIsProject && !strings.Contains(runSqlPath, constants.ScriptsDirName) {
			cmd.PrintErr("it seems like you're in a project,")
			cmd.PrintErrln(fmt.Sprintf("but the sql file you're referring to is not in the %s directory...", constants.ScriptsDirName))
			cmd.PrintErrln(fmt.Sprintf("tip: when in a project, all queries & exec statements must be present in the %s dirtectory", constants.ScriptsDirName))
		}

		return fmt.Errorf("")
	}

	sqlStr := ""
	if sqlStrB, fileErr := os.ReadFile(runSqlPath); fileErr != nil {
		cmd.PrintErrln(fmt.Sprintf("unexpected error occured while reading '%s'", runSqlPath))
		return err
	} else {
		sqlStr = string(sqlStrB)
	}

	cmd.Println("Running", runSqlPath)

	stmts := libUtils.GetSQLStmts(sqlStr)

	if len(stmts) > 1 {
		var err error
		if noTx {
			err = db_int.ExecuteStatements(stmts)
		} else {
			err = db_int.ExecuteStatementsTx(stmts)
		}

		if err != nil {
			cmd.PrintErrln("Error occured while running SQL")
			return err
		}

		cmd.Println("SQL ran successfully.")
		return nil
	} else if len(stmts) == 0 {
		cmd.PrintErrln("found 0 statements to query/execute in", sqlFileName)
		return fmt.Errorf("no statements found to execute")
	}

	return runQuery(cmd, stmts[0])
}

func runQuery(cmd *cobra.Command, query string) error {
	var wg sync.WaitGroup

	if len(query) == 0 {
		cmd.PrintErrln("WARNING: could not query, recieved empty string")
		return nil
	}

	query = strings.ToLower(strings.TrimRight(query, ";"))

	results, err := db_int.GetRows(query)
	if err != nil {
		return err
	}

	if len(results.Columns) == 0 {
		cmd.Println("Query ran successfully.")
		cmd.Println()
		return nil
	}

	cmd.Println("ROW COUNT:", results.RowCount)
	cmd.Println()

	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		cmd.Println("ERROR: couldn't get the terminal size, terminal table printing might not be ideal...")
		w = math.MaxInt32
	}

	wg.Add(1)
	defer wg.Wait()
	go (func() {
		defer wg.Done()

		if len(outFileFlag) > 0 {
			csvFile := lib.GetCsvString(results)
			outFilePath, err := libUtils.GetAbsolutePathFor(outFileFlag)
			if err != nil {
				cmd.Println("unexpected error occured while writing CSV file", err)
				return
			}

			fileErr := os.WriteFile(outFilePath, []byte(csvFile), 0644)
			if fileErr != nil {
				cmd.Println("unexpected error occured while writing CSV file", fileErr)
				return
			}

			cmd.Println("CSV output saved to file:", outFilePath)
		}
	})()

	formattedOutput := lib.GetFormattedColumns(results)
	outputW := strings.Index(formattedOutput, fmt.Sprintln()) + 1

	if len(outFileFlag) == 0 {
		if outputW <= w && results.RowCount < 100 {
			cmd.Println(formattedOutput)
		} else {
			tmpDir, tmpDirErr := libUtils.FindTmpDirPath()
			if tmpDirErr != nil {
				return err
			}

			tmpFilePath := path.Join(tmpDir, constants.TextQueryOutput)
			err := os.WriteFile(tmpFilePath, []byte(formattedOutput), 0644)
			if err != nil {
				return err
			}

			cmd.Println("Formatted output too long to display here. See tmp output file:", tmpFilePath)
		}
	}

	return nil
}

func run(cmd *cobra.Command, args []string) {
	err := cliUtils.TmpSwitchSuitableConn(cmd, func(connConfig types.ConnConfig, usingRemoteConnConfig bool) error {
		if len(args) > 0 {
			return runFile(cmd, args[0])
		}

		if queryFlag != "" {
			return runQuery(cmd, queryFlag)
		}

		cmd.Println("starting SQL shell...")
		cmd.Println("special commands: '.tables', '.exit'")
		cmd.Println()

		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("> ")

			query, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return err
			}

			query = strings.Trim(query, " ")
			query = strings.Trim(query, "\n")
			query = strings.ToLower(query)

			if query == ".tables" {
				tables, err := db_int.ListTablesInDb()
				if err != nil {
					cmd.PrintErrln("Unexpectedly, coudln't read list of tables", err)
					return err
				}

				cmd.Println("Listing tables in database:", connConfig.Database)
				for i, table := range tables {
					cmd.Printf("%d. %s.%s\n", i+1, table.Schema, table.Name)
				}
				cmd.Println()

				continue
			}

			if query == ".exit" {
				return nil
			}

			qErr := runQuery(cmd, query)
			if qErr != nil {
				cmd.PrintErrln(qErr)
			}
		}
	})
	if err != nil {
		cmd.PrintErrln("error occured:", err)
	}
}

func Init() *cobra.Command {
	cliUtils.AddRemoteFlags(cmd, &remoteFlag, &remoteNameFlag)

	cmd.Flags().BoolVar(&noTx, "no-tx", false, "when running a SQL file, by default queries are run in a transaction block, use this to disable this behaviour")
	cmd.Flags().StringVarP(&queryFlag, "query", "q", "", "enter text here to execute a one-off query")
	cmd.Flags().StringVarP(&outFileFlag, "output", "o", "", "specify CSV-formatted output file path here. supplied along with '-q'")

	return cmd
}
