package main

import (
	bc "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/blockchain"
	. "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/common"
	dbm "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/db"
	sm "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/state"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/types"
)

func main() {
	// Get BlockStore
	blockStoreDB := dbm.GetDB("blockstore")
	blockStore := bc.NewBlockStore(blockStoreDB)

	// Get State
	stateDB := dbm.GetDB("state")
	state := sm.LoadState(stateDB)

	// replay blocks on the state
	var block, nextBlock *types.Block
	if state.LastBlockHeight < blockStore.Height()-1 {
		for i := 1; i < blockStore.Height()-state.LastBlockHeight; i++ {
			block = blockStore.LoadBlock(state.LastBlockHeight + i)
			nextBlock = blockStore.LoadBlock(state.LastBlockHeight + i + 1)
			parts := block.MakePartSet()
			err := sm.ExecBlock(state, block, parts.Header())
			if err != nil {
				// TODO This is bad, are we zombie?
				PanicQ(Fmt("Failed to process committed block: %v", err))
			}
			state.Save()
			block = nextBlock
		}
	}
}
