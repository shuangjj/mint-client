package utils

import (
	cfg "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/config"
	tmcfg "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/config/tendermint"
)

func init() {
	cfg.ApplyConfig(tmcfg.GetConfig(""))
}
