package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"

	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/account"
	cclient "github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/rpc/core_client"
	"github.com/shuangjj/mint-client/Godeps/_workspace/src/github.com/shuangjj/tendermint/types"

	//cfg "github.com/shuangjj/tendermint/config"
	//tmcfg "github.com/shuangjj/tendermint/config/tendermint"
)

var (
	rpcAddr     = "pinkpenguin:46657"
	requestAddr = "http://" + rpcAddr + "/"

	PRIV_KEYS = []string{
		"9D25238C2E2221E8650ECF23A34458F5C3E5A90192EB5EDA6C4860FDCDB31508F6C79CF0CB9D66B677988BCB9B8EADD9A091CD465A60542A8AB85476256DBA92",
		"54A08F8CC74A6CB4BA8599FEEB6275E933429549362F040A31BB222DC28C97F2E15E88C226C5AEFF0597B4E71C9FEBF620538795C34CCCEB13D3CFECA8F6157B",
		"741DED293354D1BB27700559C0B6FDC5751582CBF8F9455DCEEA85A8AC78F8C5388180AC9AAF0C9A624DC0BC397A0FC0416E1713CD9181E41E8096DF5B6686FC",
		"866D9F061C2347AD40FF5B90C1EEAB49CEFDEC9329482C8C1B965A5C839A5D0DA0B0501D148232AD06BF9C57361FE35B490E2920A7622F2840F88D4097B7ECB3",
		"E7585E8E23CA0DC234BB77878E23F57CA99CB0E6F29DA7C2843A30F582565BCF6ED22473414B8DA547F5C781ADB12FB05BC6989A28BC0FADDA86D9831306F83C",
	}

	chainID = "tendermint_testnet_5e"

	keys = makeKeys()

	client = cclient.NewClient(requestAddr, "JSONRPC")
)

func makeKeys() (keys []*account.PrivAccount) {
	for _, k := range PRIV_KEYS {
		var privKeyBytes [64]byte
		keyBytes, _ := hex.DecodeString(k)
		copy(privKeyBytes[:], keyBytes)
		keys = append(keys, account.GenPrivAccountFromPrivKeyBytes(privKeyBytes[:]))
	}
	return
}

var txList = []string{"SendTx", "CallTx", "NameTx", "BondTx", "UnbondTx", "RebondTx"}

func main() {
	/*config := tmcfg.GetConfig("")
	config.Set("network", "tendermint_testnet_5e")
	cfg.ApplyConfig(config) // Notify modules of new config
	*/

	errs := new(int)

	for *errs < 100 {

		var tx types.Tx
		// pick tx type
		txType := rand.Intn(len(txList))

		switch txList[txType] {
		case "SendTx":
			tx = randomSendTx(errs)
		case "CallTx":
			// randomCallTx()
		case "NameTx":
			// randomNameTx()
		case "BondTx":
			// randomBondTx()
		}

		if tx == nil {
			continue
		}

		fmt.Println("Tx:", tx)
		rec, err := client.BroadcastTx(tx)
		if err != nil {
			fmt.Println("Err on broadcast", err)
			*errs += 1
			continue
		}
		fmt.Printf("TxID: %X\n", rec.Receipt.TxHash)
	}
}

func pickAccountNonceAmount(errs *int) (*account.PrivAccount, int, int, int64) {
	// pick sender
	i := rand.Intn(len(keys))
	privAcc := keys[i]

	acc, err := client.GetAccount(privAcc.Address)
	if err != nil {
		fmt.Println("Err on get account:", err)
		*errs += 1
		return nil, 0, 0, 0
	}
	nonce := acc.Account.Sequence

	maxAmt := 100
	amt := int64(rand.Intn(maxAmt))
	for amt == 0 {
		amt = int64(rand.Intn(maxAmt))

	}

	return privAcc, i, nonce, amt
}

func randomSendTx(errs *int) types.Tx {
	acc, i, nonce, amt := pickAccountNonceAmount(errs)
	if acc == nil {
		return nil
	}

	tx := types.NewSendTx()
	tx.AddInputWithNonce(acc.PubKey, amt, nonce)

	// pick receiver
	j := rand.Intn(len(keys))
	for ; j == i; j = rand.Intn(len(keys)) {
	}

	tx.AddOutput(keys[j].Address, amt)
	tx.SignInput(chainID, 0, acc)
	return tx
}
