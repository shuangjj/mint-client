package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/common/go/common"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/common/go/log"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/spf13/cobra"
)

var (
	DefaultKeyDaemonHost = "localhost"
	DefaultKeyDaemonPort = "4767"
	DefaultKeyDaemonAddr = DefaultKeyDaemonHost + ":" + DefaultKeyDaemonPort

	DefaultNodeRPCHost = "localhost"
	DefaultNodeRPCPort = "46657"
	DefaultNodeRPCAddr = DefaultNodeRPCHost + ":" + DefaultNodeRPCPort

	DefaultPubKey  string
	DefaultChainID string
)

// override the hardcoded defaults with env variables if they're set
func init() {
	signAddr := os.Getenv("MINTX_SIGN_ADDR")
	if signAddr != "" {
		DefaultKeyDaemonAddr = signAddr
	}

	nodeAddr := os.Getenv("MINTX_NODE_ADDR")
	if nodeAddr != "" {
		DefaultNodeRPCAddr = nodeAddr
	}

	pubKey := os.Getenv("MINTX_PUBKEY")
	if pubKey != "" {
		DefaultPubKey = pubKey
	}

	chainID := os.Getenv("MINTX_CHAINID")
	if chainID != "" {
		DefaultChainID = chainID
	}
}

var (
	signAddrFlag string
	nodeAddrFlag string
	pubkeyFlag   string
	addrFlag     string
	chainidFlag  string

	signFlag      bool
	broadcastFlag bool
	waitFlag      bool
	verboseFlag   bool
	debugFlag     bool

	// some of these are strings rather than flags because the `core`
	// functions have a pure string interface so they work nicely from http
	amtFlag      string
	nonceFlag    string
	nameFlag     string
	dataFlag     string
	dataFileFlag string
	toFlag       string
	feeFlag      string
	gasFlag      string
	unbondtoFlag string
	heightFlag   string

)

func main() {

	// these are defined in here so we can update the
	// defaults with env variables first

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "print the mintx version",
		Long:  "print the mintx version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("0.2.0") // upgrade cli to cobra
		},
	}

	var sendCmd = &cobra.Command{
		Use:   "send",
		Short: "mintx send --amt <amt> --to <addr>",
		Long:  "mintx send --amt <amt> --to <addr>",
		Run:   cliSend,
	}
	sendCmd.Flags().StringVarP(&amtFlag, "amt", "a", "", "specify an amount")
	sendCmd.Flags().StringVarP(&toFlag, "to", "t", "", "specify an address to send to")

	var nameCmd = &cobra.Command{
		Use:   "name",
		Short: "mintx name --amt <amt> --name <name> --data <data>",
		Long:  "mintx name --amt <amt> --name <name> --data <data>",
		Run:   cliName,
	}
	nameCmd.Flags().StringVarP(&amtFlag, "amt", "a", "", "specify an amount")
	nameCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "specify a name")
	nameCmd.Flags().StringVarP(&dataFlag, "data", "d", "", "specify some data")
	nameCmd.Flags().StringVarP(&dataFileFlag, "data-file", "", "", "specify a file with some data")
	nameCmd.Flags().StringVarP(&feeFlag, "fee", "f", "", "specify the fee to send")

	var callCmd = &cobra.Command{
		Use:   "call",
		Short: "mintx call --amt <amt> --fee <fee> --gas <gas> --to <contract addr> --data <data>",
		Long:  "mintx call --amt <amt> --fee <fee> --gas <gas> --to <contract addr> --data <data>",
		Run:   cliCall,
	}
	callCmd.Flags().StringVarP(&amtFlag, "amt", "a", "", "specify an amount")
	callCmd.Flags().StringVarP(&toFlag, "to", "t", "", "specify an address to send to")
	callCmd.Flags().StringVarP(&dataFlag, "data", "d", "", "specify some data")
	callCmd.Flags().StringVarP(&feeFlag, "fee", "f", "", "specify the fee to send")
	callCmd.Flags().StringVarP(&gasFlag, "gas", "g", "", "specify the gas limit for a CallTx")

	var bondCmd = &cobra.Command{
		Use:   "bond",
		Short: "mintx bond --pubkey <pubkey> --amt <amt> --unbond-to <address>",
		Long:  "mintx bond --pubkey <pubkey> --amt <amt> --unbond-to <address>",
		Run:   cliBond,
	}
	bondCmd.Flags().StringVarP(&amtFlag, "amt", "a", "", "specify an amount")
	bondCmd.Flags().StringVarP(&unbondtoFlag, "to", "t", "", "specify an address to unbond to")

	var unbondCmd = &cobra.Command{
		Use:   "unbond",
		Short: "mintx unbond --addr <address> --height <block_height>",
		Long:  "mintx unbond --addr <address> --height <block_height>",
		Run:   cliUnbond,
	}
	unbondCmd.Flags().StringVarP(&addrFlag, "addr", "a", "", "specify an address")
	unbondCmd.Flags().StringVarP(&heightFlag, "height", "h", "", "specify a height to unbond at")

	var rebondCmd = &cobra.Command{
		Use:   "rebond",
		Short: "mintx rebond --addr <address> --height <block_height>",
		Long:  "mintx rebond --addr <address> --height <block_height>",
		Run:   cliRebond,
	}
	rebondCmd.Flags().StringVarP(&addrFlag, "addr", "a", "", "specify an address")
	rebondCmd.Flags().StringVarP(&heightFlag, "height", "h", "", "specify a height to unbond at")

	var permissionsCmd = &cobra.Command{
		Use:   "permission",
		Short: "mintx perm <function name> <args ...>",
		Long:  "mintx perm <function name> <args ...>",
		Run:   cliPermissions,
	}
	permissionsCmd.Flags().StringVarP(&addrFlag, "addr", "a", "", "specify an address")
	permissionsCmd.Flags().StringVarP(&heightFlag, "height", "n", "", "specify a height to unbond at")

	var rootCmd = &cobra.Command{
		Use:               "mintx",
		Short:             "craft, sign, and broadcast tendermint transactions",
		Long:              "craft, sign, and broadcast tendermint transactions",
		PersistentPreRun:  before,
		PersistentPostRun: after,
	}

	rootCmd.PersistentFlags().StringVarP(&signAddrFlag, "sign-addr", "", DefaultKeyDaemonAddr, "set eris-keys daemon address (defaults to $MINTX_SIGN_ADDR)")
	rootCmd.PersistentFlags().StringVarP(&nodeAddrFlag, "node-addr", "", DefaultNodeRPCAddr, "set the tendermint rpc server address (defaults to $MINTX_NODE_ADDR)")
	rootCmd.PersistentFlags().StringVarP(&pubkeyFlag, "pubkey", "", DefaultPubKey, "specify the pubkey (defaults to $MINTX_PUBKEY)")
	rootCmd.PersistentFlags().StringVarP(&addrFlag, "addr", "", "", "specify the address (from which the pubkey can be fetch from eris-keys)")
	rootCmd.PersistentFlags().StringVarP(&chainidFlag, "chainID", "", DefaultChainID, "specify the pubkey (defaults to $MINTX_CHAINID)")
	rootCmd.PersistentFlags().StringVarP(&nonceFlag, "nonce", "", "", "specify the nonce to use for the transaction (should equal the sender account's nonce + 1)")

	rootCmd.PersistentFlags().BoolVarP(&signFlag, "sign", "s", false, "sign the transaction using the eris-keys daemon")
	rootCmd.PersistentFlags().BoolVarP(&broadcastFlag, "broadcast", "b", false, "broadcast the transaction to the blockchain")
	rootCmd.PersistentFlags().BoolVarP(&waitFlag, "wait", "w", false, "wait for the transaction to be committed in a block")

	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose log level")
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "debug log level")

	rootCmd.AddCommand(versionCmd, sendCmd, callCmd, nameCmd, bondCmd, unbondCmd, rebondCmd, permissionsCmd)
	common.IfExit(rootCmd.Execute())
}

func before(cmd *cobra.Command, args []string) {
	config.Set("chain_id", chainidFlag)
	if debugFlag {
		log.SetLoggers(log.LogLevelDebug, os.Stdout, os.Stderr)
	} else if verboseFlag {
		log.SetLoggers(log.LogLevelInfo, os.Stdout, os.Stderr)
	} else {
		log.SetLoggers(log.LogLevelWarn, os.Stdout, os.Stderr)
	}

	if !strings.HasPrefix(nodeAddrFlag, "http://") {
		nodeAddrFlag = "http://" + nodeAddrFlag
	}
	if !strings.HasSuffix(nodeAddrFlag, "/") {
		nodeAddrFlag += "/"
	}

	if !strings.HasPrefix(signAddrFlag, "http://") {
		signAddrFlag = "http://" + signAddrFlag
	}
}

func after(cmd *cobra.Command, args []string) {
	log.Flush()
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
