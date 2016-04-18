package state

import (
	acm "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/account"
	. "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/common"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/vm"
)

type AccountGetter interface {
	GetAccount(addr []byte) *acm.Account
}

type VMAccountState interface {
	GetAccount(addr Word256) *vm.Account
	UpdateAccount(acc *vm.Account)
	RemoveAccount(acc *vm.Account)
	CreateAccount(creator *vm.Account) *vm.Account
}
