package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/common/go/common"

	acm "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/account"
	dbm "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/db"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/merkle"
	sm "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/state"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/types"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/wire"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/spf13/cobra"
)

//------------------------------------------------------------------------------
// core dump/restore functions

// dump the latest state to json
func CoreDump(dumpval bool) []byte {
	// Get State
	stateDB := dbm.GetDB("state")
	st := sm.LoadState(stateDB)
	if st == nil {
		Exit(fmt.Errorf("Error: state loaded from %s is nil!", config.GetString("db_dir")))
	}

	stJ := new(State)

	//default is true, flag omits vals from dump
	if dumpval {
		stJ.BondedValidators = st.BondedValidators
		stJ.LastBondedValidators = st.LastBondedValidators
		stJ.UnbondingValidators = st.UnbondingValidators
	}
	// iterate through accounts tree
	// track storage roots as we go
	storageRoots := [][]byte{}
	st.GetAccounts().Iterate(func(key interface{}, value interface{}) (stopped bool) {
		acc := value.(*acm.Account)
		stJ.Accounts = append(stJ.Accounts, acc)
		storageRoots = append(storageRoots, acc.StorageRoot)
		return false
	})

	// grab all storage
	for i, root := range storageRoots {
		if len(root) == 0 {
			continue
		}
		accStorage := &AccountStorage{Address: stJ.Accounts[i].Address}

		storage := merkle.NewIAVLTree(wire.BasicCodec, wire.BasicCodec, 1024, stateDB)
		storage.Load(root)
		storage.Iterate(func(key interface{}, value interface{}) (stopped bool) {
			k, v := key.([]byte), value.([]byte)
			accStorage.Storage = append(accStorage.Storage, &Storage{k, v})
			return false
		})
		stJ.AccountsStorage = append(stJ.AccountsStorage, accStorage)
	}

	// get all validator infos
	if dumpval {
		st.GetValidatorInfos().Iterate(func(key interface{}, value interface{}) (stopped bool) {
			vi := value.(*types.ValidatorInfo)
			stJ.ValidatorInfos = append(stJ.ValidatorInfos, vi)
			return false
		})
	}
	// get all name entries
	st.GetNames().Iterate(func(key interface{}, value interface{}) (stopped bool) {
		name := value.(*types.NameRegEntry)
		stJ.NameReg = append(stJ.NameReg, name)
		return false
	})

	w, n, err := new(bytes.Buffer), new(int64), new(error)
	wire.WriteJSON(stJ, w, n, err)

	IfExit(*err)
	w2 := new(bytes.Buffer)
	json.Indent(w2, w.Bytes(), "", "\t")
	return w2.Bytes()
}

// restore state from json blob
// set tendermint config before calling
func CoreRestore(chainID string, jsonBytes []byte) {
	var stJ State
	var err error
	wire.ReadJSON(&stJ, jsonBytes, &err)
	IfExit(err)

	st := new(sm.State)

	st.ChainID = chainID
	st.BondedValidators = stJ.BondedValidators
	st.LastBondedValidators = stJ.LastBondedValidators
	st.UnbondingValidators = stJ.UnbondingValidators

	stateDB := dbm.GetDB("state")

	// fill the accounts tree
	accounts := merkle.NewIAVLTree(wire.BasicCodec, acm.AccountCodec, 1000, stateDB)
	for _, account := range stJ.Accounts {
		accounts.Set(account.Address, account.Copy())
	}

	// fill the storage tree for each contract
	for _, accStorage := range stJ.AccountsStorage {
		st := merkle.NewIAVLTree(wire.BasicCodec, wire.BasicCodec, 1024, stateDB)
		for _, accSt := range accStorage.Storage {
			set := st.Set(accSt.Key, accSt.Value)
			if !set {
				panic("failed to update storage tree")
			}
		}
		// TODO: sanity check vs acc.StorageRoot

		st.Save()
	}

	valInfos := merkle.NewIAVLTree(wire.BasicCodec, types.ValidatorInfoCodec, 0, stateDB)
	for _, valInfo := range stJ.ValidatorInfos {
		valInfos.Set(valInfo.Address, valInfo)
	}

	nameReg := merkle.NewIAVLTree(wire.BasicCodec, sm.NameRegCodec, 0, stateDB)
	for _, entry := range stJ.NameReg {
		nameReg.Set(entry.Name, entry)
	}

	// persists accounts/valInfos/nameReg trees
	st.SetAccounts(accounts)
	st.SetValidatorInfos(valInfos)
	st.SetNameReg(nameReg)
	st.SetDB(stateDB)
	st.Save()
}

//------------------------------------------------------------------------------
// cli wrappers

func cliRestore(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		Exit(fmt.Errorf("Enter the chain id"))
	}
	chainID := args[0]

	var err error
	var b []byte

	if IPFShash == "" {
		fi, _ := os.Stdin.Stat()
		if fi.Size() == 0 {
			Exit(fmt.Errorf("Please pass data to restore on Stdin or specify IPFS hash with --ipfs=\"[hash]\""))
		}
		b, err = ioutil.ReadAll(os.Stdin)
		IfExit(err)
	} else {
		url := composeIPFSUrl(HostFlag, ApiFlag)

		w := bytes.NewBuffer([]byte{})
		w.Write([]byte("Reading file from IPFS. Hash =>\t" + IPFShash + "\n"))
		b, err = IPFSCat(url, IPFShash, ApiFlag, w)
		IfExit(err)
	}

	CoreRestore(chainID, b)

	stateDB := dbm.GetDB("state")
	newState := sm.LoadState(stateDB)
	fmt.Printf("State hash: %X\n", newState.Hash())

}

//TODO stop node / copy issue #18
func cliDump(cmd *cobra.Command, args []string) {
	state := CoreDump(DumpValidatorsFlag)

	if !DumpToIPFSFlag {
		fmt.Println(string(state))
	} else {
		url := composeIPFSUrl(HostFlag, ApiFlag)
		hash, err := IPFSUpload(url, state, bytes.NewBuffer([]byte{}))
		if err != nil {
			fmt.Println("problem sending to IPFS: %v", err)
		}
		fmt.Println(hash)
	}
}

//------------------------------------------------------------------------------
// types

type State struct {
	BondedValidators     *types.ValidatorSet    `json:"bonded_validators"`
	LastBondedValidators *types.ValidatorSet    `json:"last_bonded_validators"`
	UnbondingValidators  *types.ValidatorSet    `json:"unbonding_validators"`
	Accounts             []*acm.Account         `json:"accounts"`
	AccountsStorage      []*AccountStorage      `json:"accounts_storage"`
	ValidatorInfos       []*types.ValidatorInfo `json:"validator_infos"`
	NameReg              []*types.NameRegEntry  `json:"namereg"`
}

type AccountStorage struct {
	Address []byte     `json:"address"`
	Storage []*Storage `json:"storage"`
}

type Storage struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}
