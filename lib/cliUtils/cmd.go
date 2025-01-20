package cliUtils

import (
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/types"
	"github.com/spf13/cobra"
)

func AddRemoteFlags(cmd *cobra.Command, remoteFlag *bool, remoteNameFlag *string) {
	cmd.PersistentFlags().BoolVarP(remoteFlag, "remote", "r", false, "use a remote origin. use 'remote-name ...' if multiple remote origins are present")
	cmd.PersistentFlags().StringVar(remoteNameFlag, "remote-name", "", "use a named remote origin.")
}

// check for remoteFlag & remoteNameFlag, switch to the remote db if specified
// callback function is passed in with connConfig and whether the connConfig is a remote config or not
func TmpSwitchSuitableConn(cmd *cobra.Command, fn func(connConfig types.ConnConfig, usingRemoteConnConfig bool) error) error {
	remoteFlag, remoteFlagErr := cmd.Flags().GetBool("remote")
	if remoteFlagErr != nil {
		cmd.Println("WARNING:", remoteFlagErr)
	}
	remoteNameFlag, remoteNameFlagErr := cmd.Flags().GetString("remote-name")
	if remoteNameFlagErr != nil {
		cmd.Println("WARNING:", remoteNameFlagErr)
	}
	usingRemoteConnConfig := remoteFlag || remoteNameFlag != ""

	connConfig, connConfigErr := lib.GetSuitableConnConfig(&globals.CliConfig, remoteFlag, remoteNameFlag)
	if connConfigErr != nil {
		return connConfigErr
	}

	return lib.TmpSwitchConn(connConfig, func() error {
		return fn(connConfig, usingRemoteConnConfig)
	})
}
