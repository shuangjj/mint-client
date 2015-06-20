package main

import (
	"fmt"
	"github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/spf13/cobra"
	"os"
	"os/user"
	"path"
)

var (
	usr, _ = user.Current()

	DefaultKeyStore = path.Join(usr.HomeDir, ".eris", "keys")
)

func main() {
	var stringsToIntsCmd = &cobra.Command{
		Use:   "int",
		Short: "Convert list of permissions to PermFlag and SetBit",
		Long:  "Example: mintperms int call:0 send:1 name:1",
		Run:   cliStringsToInts,
	}
	var intsToStringsCmd = &cobra.Command{
		Use:   "string",
		Short: "Convert PermFlag and SetBit integers to strings",
		Long:  "Example: mintperms string 2 6",
		Run:   cliIntsToStrings,
	}
	var bbpbCmd = &cobra.Command{
		Use:   "bbpb",
		Short: "Print the permissions for a BBPB",
		Run:   cliBBPB,
	}
	var allCmd = &cobra.Command{
		Use:   "all",
		Short: "Print the PermFlag and SetBit for all permissions on and set",
		Run:   cliAll,
	}

	var rootCmd = &cobra.Command{Use: "mintperms"}
	rootCmd.AddCommand(stringsToIntsCmd, intsToStringsCmd, bbpbCmd, allCmd)
	rootCmd.Execute()
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func ifExit(err error) {
	if err != nil {
		exit(err)
	}
}
