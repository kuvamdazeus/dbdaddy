package types

import "github.com/spf13/cobra"

type CobraCmdFn = func(cmd *cobra.Command, args []string)

type MiddlewareFunc = func(fn CobraCmdFn) CobraCmdFn