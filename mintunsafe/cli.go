package main

import (
	"fmt"
	"os"

	. "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/common/go/common"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/types"

	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/spf13/cobra"
)

//------------------------------------------------------------------------------

// NOTE: this is totally unsafe.
// it's only suitable for testnets.
func cliResetPriv(cmd *cobra.Command, args []string) {
	// Get PrivValidator
	pvFile := config.GetString("priv_validator_file")
	if _, err := os.Stat(pvFile); err != nil {
		Exit(err)
	}
	pV := types.LoadPrivValidator(pvFile)
	pV.LastHeight, pV.LastRound, pV.LastStep = 0, 0, 0
	pV.Save()
	fmt.Println("Reset PrivValidator", pvFile)
}
