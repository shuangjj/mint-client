package main

import (
	cfg "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/config"

	_ "github.com/shuangjj/mint-client/utils" // calls ApplyConfig
)

var config cfg.Config = nil

func init() {
	cfg.OnConfig(func(newConfig cfg.Config) {
		config = newConfig
	})
}
