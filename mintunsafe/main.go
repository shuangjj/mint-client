package main

import (
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/spf13/cobra"
)

func main() {
	var resetPrivCmd = &cobra.Command{
		Use:   "reset-priv",
		Short: "reset the priv-validator fields to 0",
		Long:  "reset the priv-validator fields to 0",
		Run:   cliResetPriv,
	}

	var rootCmd = &cobra.Command{Use: "mintunsafe"}
	rootCmd.AddCommand(resetPrivCmd)
	rootCmd.Execute()
}
