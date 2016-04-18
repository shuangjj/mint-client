package main

import (
	"fmt"

	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/spf13/cobra"
)

var (
	DirFlag string
	//AddrsFlag  string
	CsvPathFlag       string
	PubkeyFlag        string
	RootFlag          string
	NoValAccountsFlag bool
)

func main() {

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "check mintgen version",
		Run: func(cmd *cobra.Command, args []string) {
			// major: alpha
			// minor: drop pub, csvs with empty fields, no config
			// revision: csvs for validators and accounts
			fmt.Println("0.3.1")
		},
	}

	var randomCmd = &cobra.Command{
		Use:   "random",
		Short: "mintgen random [flags] <N> <chain_id>",
		Long:  "Create <N> keys and a genesis.json with corresponding validators and chain_id <name>",
		Run:   cliRandom,
	}

	var knownCmd = &cobra.Command{
		Use:   "known",
		Short: "mintgen known <chain_id> [flags] ",
		Long:  "Create a genesis.json with --pub <pub_1> <pub_2> <pub_N> or with --csv <path_to_file>, or pass a priv_validator.json on stdin. Two csv file names can be passed (comma separated) to distinguish validators and accounts.",
		Run:   cliKnown,
	}

	randomCmd.Flags().StringVarP(&DirFlag, "dir", "d", "", "Directory to save genesis and priv_validators in. Default is ~/.eris/blockchains/<chain_id>")
	randomCmd.Flags().StringVarP(&PubkeyFlag, "pub", "p", "", "pubkeys to include in accounts when generating genesis.json")
	randomCmd.Flags().StringVarP(&CsvPathFlag, "csv", "", "", "Path to .csv with the following params: (pubkey, starting balance, name, permissions, setbit")
	randomCmd.Flags().StringVarP(&RootFlag, "root", "r", "", "pubkeys to include in accounts and give root permissions to")
	randomCmd.Flags().BoolVarP(&NoValAccountsFlag, "no-val", "n", false, "don't give the validators accounts")

	knownCmd.Flags().StringVarP(&PubkeyFlag, "pub", "", "", "pubkeys to include when generating genesis.json. flag is req'd")
	knownCmd.Flags().StringVarP(&CsvPathFlag, "csv", "", "", "Path to .csv with the following params: (pubkey, starting balance, name, permissions, setbit")

	var rootCmd = &cobra.Command{
		Use:   "mintgen",
		Short: "a tool for generating tendermint genesis files",
		Long:  "a tool for generating tendermint genesis files",
	}
	rootCmd.AddCommand(randomCmd, knownCmd, versionCmd)
	rootCmd.Execute()
}
