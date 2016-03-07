package main

import (
	"fmt"
	"path"

	. "github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	cfg "github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/eris-ltd/tendermint/config"
	tmcfg "github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/eris-ltd/tendermint/config/tendermint"
	"github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/spf13/cobra"
	"os"
)

var (
	config = tmcfg.GetConfig("")

	ApiFlag            bool
	DumpToIPFSFlag     bool
	DumpValidatorsFlag bool
	DataDirFlag        string
	IPFShash           string
	HostFlag           string

//	StopNode    bool
)

func init() {
	cfg.ApplyConfig(config)
}

func main() {

	var dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump tendermint state to json files or IPFS",
		Long:  "mintdump dump > [fileName.json], mintdump dump --ipfs [fileName.json] ",
		Run:   cliDump,
	}
	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore tendermint state from json files or by IPFS hash",
		Long:  "mintdump restore [new chainID] < [path/to/file], mintdump restore [new chainID] --ipfs [hash] ",
		Run:   cliRestore,
	}
	dumpCmd.Flags().StringVarP(&DataDirFlag, "data-dir", "d", "", "Path to tendermint data directory")
	dumpCmd.Flags().BoolVarP(&DumpToIPFSFlag, "ipfs", "", false, "Dump state to IPFS as json.")
	//dumpCmd.Flags().BoolVarP(&StopNode, "stop", "s", false, "stop node if it is running. Req'd for mintdump")
	dumpCmd.Flags().BoolVarP(&DumpValidatorsFlag, "val", "", true, "Omit validators from dump with --val=false")
	//not supported
	//dumpCmd.Flags().BoolVarP(&ApiFlag, "api", "", false, "Use IPFS api. Req's ipfs daemon running locally or as an (eris) service. Gateway is default and req's neither")
	dumpCmd.Flags().StringVarP(&HostFlag, "host", "", "", "Set the host for IPFS")

	restoreCmd.Flags().StringVarP(&DataDirFlag, "data-dir", "d", "", "Path to tendermint data directory")
	restoreCmd.Flags().StringVarP(&IPFShash, "ipfs", "", "", "Restore .json from IPFS, by hash")
	restoreCmd.Flags().BoolVarP(&ApiFlag, "api", "", false, "Use IPFS api. Requires ipfs daemon running locally or as an (eris) service")
	restoreCmd.Flags().StringVarP(&HostFlag, "host", "", "", "Set the host for IPFS")

	var rootCmd = &cobra.Command{Use: "mintdump"}
	rootCmd.PersistentPreRun = before
	rootCmd.AddCommand(dumpCmd, restoreCmd)
	rootCmd.Execute()
}

func before(cmd *cobra.Command, args []string) {
	if DataDirFlag != "" {
		if _, err := os.Stat(path.Join(DataDirFlag, "state.db")); err != nil {
			Exit(fmt.Errorf("Could not find state.db folder in %s", DataDirFlag))
		}
		config.Set("db_dir", DataDirFlag)
		cfg.ApplyConfig(config)
	}
}
