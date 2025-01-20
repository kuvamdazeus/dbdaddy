package checkoutCmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/errs"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/lib/libUtils"
	"github.com/fossmedaddy/dbdaddy/middlewares"

	"github.com/spf13/cobra"
)

// flags
var (
	shouldCreateNewBranch bool
	shouldKeepItClean     bool
	shouldCopyOnlySchema  bool
)

var cmdRunFn = middlewares.Apply(run, middlewares.CheckConnection)

var cmd = &cobra.Command{
	Use:   "checkout <branchname>",
	Short: "checkout into a new/existing branch in the database",
	Args:  cobra.ExactArgs(1),
	Run:   cmdRunFn,
}

func Init() *cobra.Command {
	cmd.Flags().BoolVarP(&shouldCreateNewBranch, "new", "n", false, "create a new branch with given branch name, the contents of the current branch will be copied over schema and data both")
	cmd.Flags().BoolVarP(&shouldKeepItClean, "clean", "c", false, "the new branch created by '-n' will be independent of the current branch i.e. nothing will be copied over")
	cmd.Flags().BoolVar(&shouldCopyOnlySchema, "only-schema", false, "create a new branch with given branch name, only the schema of the current branch will be copied over")
	return cmd
}

func run(cmd *cobra.Command, args []string) {
	configDirPath, configErr := libUtils.FindConfigDirPath()
	if configErr != nil {
		cmd.PrintErrln("unexpected error occured!", configErr)
		os.Exit(1)
	}

	branchname := args[0]

	_, cwdIsProject, cwdErr := libUtils.CwdIsProject()
	if cwdErr != nil {
		cmd.PrintErrln("unexpected error occured!")
		cmd.PrintErrln(cwdErr)
		return
	}

	if cwdIsProject {
		cmd.Println("in a project, can't switch branches!")
		cmd.Println("a project is attached to ONLY ONE DATABASE")
		return
	}

	// flags validation
	if shouldKeepItClean && !shouldCreateNewBranch {
		cmd.PrintErrln("'--clean, -c' option can only be provided when creating new branches i.e. with '-n, --new' option")
		return
	}
	if shouldCopyOnlySchema && !shouldCreateNewBranch {
		cmd.PrintErrln("'--only-schema' option can only be provided when creating new branches i.e. with '-n, --new' option")
		return
	}
	if shouldCopyOnlySchema && shouldKeepItClean {
		cmd.PrintErrln("'--clean, -c' and '--only-schema' options can not be used simultaneously")
		return
	}

	if shouldCreateNewBranch {
		var err error
		if shouldKeepItClean {
			err = db_int.CreateDb(branchname)
		} else {
			err = lib.NewBranchFromCurrent(branchname, shouldCopyOnlySchema)
		}
		if err != nil {
			if errors.Is(err, errs.ErrDbAlreadyExists) {
				cmd.PrintErrf("Could not create a new database branch with name '%s' because it already exists.\n", branchname)
				return
			}

			cmd.PrintErrln("UNKNOWN ERROR OCCURED!", err)
			return
		}
	} else {
		if !db_int.DbExists(branchname) {
			cmd.PrintErrf("Database branch '%s' does not exists, run 'checkout <branchname> -n' to create a new branch\n", branchname)
			return
		}
	}

	globals.CliConfig.State.CurrentBranch = branchname
	if err := lib.WriteConfig(globals.CliConfig, configDirPath, false); err != nil {
		cmd.PrintErrln("unexpected error occured!", err)
		os.Exit(1)
	}

	cmd.Println(fmt.Sprintf("Switched to branch: %s", branchname))
}
