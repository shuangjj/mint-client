package main

import (
	. "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/common/go/log"
)

var logger *Logger

func init() {
	logger = AddLogger("mintx")
}
