package main

import (
	"encoding/hex"
	"log"
	"testing"
	"time"

	"github.com/polygonledger/node/block"
	chain "github.com/polygonledger/node/chain"
	"github.com/polygonledger/node/crypto"
	"github.com/polygonledger/node/ntcl"
)

func TestBasicCommand(t *testing.T) {

	log.Println("TestBasicCommand")

	mgr := chain.CreateManager()
	mgr.InitAccounts()

	req_msg := ntcl.EncodeMsgString(ntcl.REQ, ntcl.CMD_PING, ntcl.EMPTY_DATA)
	msg := ntcl.ParseMessage(req_msg)
	//ntchan.REQ_in <- msg
	reply_msg := HandlePing(msg)
	if reply_msg != "REP#PONG#|" {
		t.Error(reply_msg)
	}

	ntchan := ntcl.ConnNtchanStub("")
	go RequestHandlerTel(&mgr, ntchan)
	ntchan.REQ_in <- req_msg
	reply := <-ntchan.REP_out

	if reply != "REP#PONG#|" {
		t.Error(reply_msg)
	}

}

func TestBalance(t *testing.T) {

	log.Println("TestBalance")

	mgr := chain.CreateManager()
	mgr.InitAccounts()

	req_msg := ntcl.EncodeMsgString(ntcl.REQ, ntcl.CMD_BALANCE, "abc")
	msg := ntcl.ParseMessage(req_msg)

	reply_msg := HandleBalance(&mgr, msg)
	if reply_msg != "REP#BALANCE#0|" {
		t.Error(reply_msg)
	}

	//TODO with chain setup

	//log.Println(mgr.Accounts)
	ra := mgr.RandomAccount()
	mgr.SetAccount(ra, 10)

	log.Println(ra.AccountKey)
	req_msg = ntcl.EncodeMsgString(ntcl.REQ, ntcl.CMD_BALANCE, ra.AccountKey)
	msg = ntcl.ParseMessage(req_msg)

	reply_msg = HandleBalance(&mgr, msg)
	if reply_msg != "REP#BALANCE#10|" {
		t.Error(reply_msg)
	}

	log.Println(mgr.Accounts)

	//log.Println(ra)

	// b := block.Block{}
	// tx := tx.Tx{}
	// mgr.ApplyBlock(b)
}

func TestFaucetTx(t *testing.T) {

	kp := crypto.PairFromSecret("test")
	pubk := crypto.PubKeyToHex(kp.PubKey)
	addr := crypto.Address(pubk)
	//faucet_to := block.AccountFromString(addr)

	//accountJson, _ := json.Marshal(faucet_to)
	//req_msg := ntcl.EncodeMsgString(ntcl.REQ, ntcl.CMD_FAUCET, string(accountJson))
	req_msg := ntcl.EncodeMsgString(ntcl.REQ, ntcl.CMD_FAUCET, addr)
	msg := ntcl.ParseMessage(req_msg)

	mgr := chain.CreateManager()
	mgr.InitAccounts()

	reply_msg := HandleFaucet(&mgr, msg)
	if reply_msg != "REP#FAUCET#ok|" {
		t.Error(reply_msg)
	}
	chain.MakeBlock(&mgr)

	time.Sleep(1000 * time.Millisecond)

	req_msg = ntcl.EncodeMsgString(ntcl.REQ, ntcl.CMD_BALANCE, addr)
	msg = ntcl.ParseMessage(req_msg)

	reply_msg = HandleBalance(&mgr, msg)
	if reply_msg != "REP#BALANCE#10|" {
		t.Error(reply_msg)
	}

}

func TestTx(t *testing.T) {

	mgr := chain.CreateManager()
	mgr.InitAccounts()
	kp := crypto.PairFromSecret("test")
	pubk := crypto.PubKeyToHex(kp.PubKey)
	addr := crypto.Address(pubk)
	req_msg := ntcl.EncodeMsgString(ntcl.REQ, ntcl.CMD_FAUCET, addr)
	msg := ntcl.ParseMessage(req_msg)

	reply_msg := HandleFaucet(&mgr, msg)
	if reply_msg != "REP#FAUCET#ok|" {
		t.Error(reply_msg)
	}
	chain.MakeBlock(&mgr)
	time.Sleep(100 * time.Millisecond)
	req_msg = ntcl.EncodeMsgString(ntcl.REQ, ntcl.CMD_BALANCE, addr)
	msg = ntcl.ParseMessage(req_msg)

	reply_msg = HandleBalance(&mgr, msg)
	if reply_msg != "REP#BALANCE#10|" {
		t.Error(reply_msg)
	}

	sender := block.AccountFromString(addr)

	kp2 := crypto.PairFromSecret("test2")
	pubk2 := crypto.PubKeyToHex(kp2.PubKey)
	addr2 := crypto.Address(pubk2)
	recv := block.AccountFromString(addr2)

	amount := 5
	tx := block.Tx{Nonce: 1, Amount: amount, Sender: sender, Receiver: recv}
	signature := crypto.SignTx(tx, kp)
	sighex := hex.EncodeToString(signature.Serialize())
	// if sighex != "3045022100c360a962aeb6dcee880c45e5be84ee20df7169d6ab2ea5a94228e2fb16b955e5022048a411bc2a85e2aff76d8172abaced9780f03527ab8efbfc8cc380bdb40ccb7a" {
	// 	t.Error(sighex)
	// }
	tx.Signature = sighex
	tx.SenderPubkey = crypto.PubKeyToHex(kp.PubKey)

	verified := crypto.VerifyTxSig(tx)

	if !verified {
		t.Error("not verified")
	}

	valid := chain.TxValid(&mgr, tx)

	if !valid {
		t.Error("not valid")
	}

	// txJson, _ := json.Marshal(tx)
	// req_msg = ntcl.EncodeMessageTx(txJson)
	// msg = ntcl.ParseMessage(req_msg)

	// reply_msg = HandleTx(&mgr, msg)
	// //TODO!
	// if reply_msg == "" {
	// 	// 	t.Error(reply_msg)
	// }
}

func TestRanaccount(t *testing.T) {
	// REQ#RANACC#|
	// REP#RANACC#{"AccountKey":"Pe2c32a4f7e8b"}|
	// REQ#BALANCE#Pe2c32a4f7e8b/
	// REP#BALANCE#0|
	// REQ#BALANCE#Pe2c32a4f7e8b|
	// REP#BALANCE#20|
}

func TestGenesis(t *testing.T) {

	// genBlock := chain.MakeGenesisBlock()
	// mgr.ApplyBlock(genBlock)
	// //chain.SetAccount()

	// for k, v := range mgr.Accounts {
	// 	fmt.Println(k, v)
	// 	if !mgr.IsTreasury(k) {
	// 		if v != 20 {
	// 			t.Error("...")
	// 		}
	// 	} else {
	// 		if v != 200 {
	// 			t.Error("...")
	// 		}
	// 	}

}
