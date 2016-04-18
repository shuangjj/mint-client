package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"testing"

	cfg "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/config"
	dbm "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/db"
	ptypes "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/permission/types"
	sm "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/state"
)

var TestDir = path.Join(os.Getenv("GOPATH"), "src", "github.com", "shuangjj", "mint-client", "mintdump", "test")

/*
// XXX: Keeping leveldb files in version control is not nice ....
func TestDumpRestore(t *testing.T) {
	// load a test dir
	config.Set("db_dir", path.Join(TestDir, "data1"))
	cfg.ApplyConfig(config) // Notify modules of new config
	stateDB := dbm.GetDB("state")
	st := sm.LoadState(stateDB)
	stHash := st.Hash()
	dump := CoreDump()

	// restore to a memdir
	config.Set("db_backend", "memdb")
	cfg.ApplyConfig(config) // Notify modules of new config
	CoreRestore("", dump)

	stateDB = dbm.GetDB("state")
	st = sm.LoadState(stateDB)
	if bytes.Compare(stHash, st.Hash()) != 0 {
		t.Fatalf("State hash mismatch. Got %X, expected %X", st.Hash(), stHash)
	}
}
*/

func TestRestoreDump(t *testing.T) {
	b1, err := ioutil.ReadFile(path.Join(TestDir, "data1.json")) // with validators
	if err != nil {
		t.Fatal(err)
	}
	b2, err := ioutil.ReadFile(path.Join(TestDir, "data2.json")) // without
	if err != nil {
		t.Fatal(err)
	}
	b1 = bytes.Trim(b1, "\n")
	b2 = bytes.Trim(b2, "\n")

	// restore to a memdir
	config.Set("db_backend", "memdb")
	cfg.ApplyConfig(config) // Notify modules of new config
	CoreRestore("", b1)

	stateDB := dbm.GetDB("state")
	st := sm.LoadState(stateDB)
	acc := st.GetAccount(ptypes.GlobalPermissionsAddress)
	fmt.Println(acc)

	dump1 := CoreDump(true) // with validators

	if bytes.Compare(b1, dump1) != 0 {
		ld, lb := len(dump1), len(b1)
		max := int(math.Max(float64(ld), float64(lb)))
		n := 100
		for i := 0; i < max/n; i++ {
			dd := dump1[i*n : (i+1)*n]
			bb := b1[i*n : (i+1)*n]
			if bytes.Compare(dd, bb) != 0 {
				t.Fatalf("Error in dumps! Got \n\n\n\n %s \n\n\n\n Expected \n\n\n\n %s", dd, bb)
			}
		}
	}

	CoreRestore("", b2)
	dump2 := CoreDump(false) //without validators
	if bytes.Compare(b2, dump2) != 0 {
		ld, lb := len(dump2), len(b2)
		max := int(math.Max(float64(ld), float64(lb)))
		n := 100
		for i := 0; i < max/n; i++ {
			dd := dump2[i*n : (i+1)*n]
			bb := b2[i*n : (i+1)*n]
			if bytes.Compare(dd, bb) != 0 {
				t.Fatalf("Error in dumps! Got \n\n\n\n %s \n\n\n\n Expected \n\n\n\n %s", dd, bb)
			}
		}
	}
}
