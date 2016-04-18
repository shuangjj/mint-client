package main

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/common/go/common"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/spf13/cobra"
)

func cliConfig(cmd *cobra.Command, args []string) {

	checkFlags(nodeAddr, seeds, db_backend, log_level, rpcAddr, fast_sync)

	//XXX this formating necessary for properly formatted stdout
	var Config = fmt.Sprintf(
		`# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

moniker = "%s"
skip_upnp = %t
node_laddr = "%s"
seeds = "%s"
fast_sync = %t
db_backend = "%s"
log_level = "%s"
rpc_laddr = "%s"`, moniker, skip_upnp, nodeAddr, seeds, fast_sync, db_backend, log_level, rpcAddr)

	fmt.Printf("%v\n", Config)

}

func validateAddress(name, address string) {
	if address != "" {
		p := strings.Split(address, ":")
		if len(p) != 2 {
			Exit(fmt.Errorf("--%s should be <host>:<port>", name))
		}
		if _, err := strconv.Atoi(p[1]); err != nil {
			Exit(fmt.Errorf("specified --%s port must be an integer", name))
		}
	}
}

func validateValueInList(name, value string, acceptedValues []string) {
	for _, v := range acceptedValues {
		if value == v {
			return
		}
	}
	Exit(fmt.Errorf("--%s must be one of %v", name, acceptedValues))
}

func checkFlags(node, seeds, db, log, rpc string, sync bool) {
	validateAddress("p2p", node)
	validateAddress("rpc", rpc)
	validateAddress("seeds", seeds)

	validateValueInList("db", db, []string{"memdb", "leveldb"})
	validateValueInList("log", log, []string{"error", "warn", "notice", "info", "debug"})
}
