package main

//client based application to request information from peers

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/polygonledger/node/block"
	"github.com/polygonledger/node/crypto"
	"github.com/polygonledger/node/ntcl"
	"github.com/polygonledger/node/ntwk"
)

var Peers []ntwk.Peer

type Configuration struct {
	PeerAddresses []string
	NodePort      int
	WebPort       int
}

func addPeerOut(p ntwk.Peer) {
	Peers = append(Peers, p)
	log.Println("peers now", Peers)
}

func runningtime(s string) (string, time.Time) {
	log.Println("Start:	", s)
	return s, time.Now()
}

func track(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("End measure time:", s, "took", endTime.Sub(startTime))
}

//
func MakeRandomTx(peer ntwk.Peer) error {
	//make a random transaction by requesting random account from node
	//get random account

	req_msg := ntwk.EncodeMsgString(ntwk.REQ, ntwk.CMD_RANDOM_ACCOUNT, "emptydata")

	response := ntwk.RequestReplyChan(req_msg, peer.Req_chan, peer.Rep_chan)

	var a block.Account
	dataBytes := []byte(response.Data)
	if err := json.Unmarshal(dataBytes, &a); err != nil {
		panic(err)
	}
	log.Print(" account key ", a.AccountKey)

	//use this random account to send coins from

	//send Tx
	testTx := ntwk.RandomTx(a)
	txJson, _ := json.Marshal(testTx)
	log.Println("txJson ", txJson)

	req_msg = ntwk.EncodeMessageTx(txJson)

	response = ntwk.RequestReplyChan(req_msg, peer.Req_chan, peer.Rep_chan)
	log.Print("response msg ", response)

	return nil
}

func CreateTx(peer ntwk.Peer) {
	// keypair := crypto.PairFromSecret("test")
	// var tx block.Tx
	// s := block.AccountFromString("Pa033f6528cc1")
	// r := s //TODO
	// tx = block.Tx{Nonce: 0, Amount: 0, Sender: s, Receiver: r}
	// signature := crypto.SignTx(tx, keypair)
	// sighex := hex.EncodeToString(signature.Serialize())
	// tx.Signature = sighex
	// tx.SenderPubkey = crypto.PubKeyToHex(keypair.PubKey)

}

func PushTx(peer ntwk.Peer) error {

	dat, _ := ioutil.ReadFile("tx.json")
	var tx block.Tx

	if err := json.Unmarshal(dat, &tx); err != nil {
		panic(err)
	}

	//send Tx
	txJson, _ := json.Marshal(tx)
	log.Println("txJson ", string(txJson))

	req_msg := ntwk.EncodeMessageTx(txJson)
	response := ntwk.RequestReplyChan(req_msg, peer.Req_chan, peer.Rep_chan)
	log.Print("reply msg ", response)

	return nil
}

func Getbalance(ntchan ntcl.Ntchan) error {
	//user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter address: ")
	addr, _ := reader.ReadString('\n')
	addr = strings.Trim(addr, string('\n'))

	txJson, _ := json.Marshal(block.Account{AccountKey: addr})
	req_msg := ntwk.EncodeMsgString(ntwk.REQ, ntwk.CMD_BALANCE, string(txJson))
	//response := ntwk.RequestReplyChan(req_msg, peer.Req_chan, peer.Rep_chan)
	ntchan.REQ_out <- req_msg

	time.Sleep(1000 * time.Millisecond)

	response := <-ntchan.REP_in
	resonse_msg := ntwk.ParseMessage(response)

	log.Println("response ", response)
	var balance int
	if err := json.Unmarshal(resonse_msg.Data, &balance); err != nil {
		panic(err)
	}
	log.Println("balance of account ", balance)

	return nil
}

func Getblockheight(peer ntwk.Peer) error {
	req_msg := ntwk.EncodeMsgString(ntwk.REQ, ntwk.CMD_BLOCKHEIGHT, "")
	response := ntwk.RequestReplyChan(req_msg, peer.Req_chan, peer.Rep_chan)

	var blockheight int
	if err := json.Unmarshal(response.Data, &blockheight); err != nil {
		panic(err)
	}
	log.Println("blockheight ", blockheight)

	return nil
}

func Gettxpool(peer ntwk.Peer) error {
	req_msg := ntwk.EncodeMsgString(ntwk.REQ, ntwk.CMD_GETTXPOOL, "")
	log.Println("> ", req_msg)
	resp := ntwk.RequestReplyChan(req_msg, peer.Req_chan, peer.Rep_chan)

	log.Println("rcvmsg ", resp)
	log.Println("data ", resp.Data)

	var txp []block.Tx
	if err := json.Unmarshal(resp.Data, &txp); err != nil {
		panic(err)
	}
	log.Println("txp ", txp)

	return nil
}

func GetFaucet(peer ntwk.Peer) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter address: ")
	addr, _ := reader.ReadString('\n')
	addr = strings.Trim(addr, string('\n'))

	accountJson, _ := json.Marshal(block.Account{AccountKey: addr})
	req_msg := ntwk.EncodeMsgString(ntwk.REQ, ntwk.CMD_FAUCET, string(accountJson))
	resp := ntwk.RequestReplyChan(req_msg, peer.Req_chan, peer.Rep_chan)
	log.Println("resp ", resp)

	return nil
}

func readdns() {
	// domain := "example.com"
	// ips, err1 := net.LookupIP(domain)
	// if err1 != nil {
	// 	fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err1)
	// 	os.Exit(1)
	// }
	// for _, ip := range ips {
	// 	fmt.Printf(domain+". IN A %s\n", ip.String())
	// }

}

//setup connection to a peer from client side for requests
//TODO old
func setupPeerClient(peer ntwk.Peer) error {

	ntchan := ntwk.OpenNtchanOut(peer.Address, peer.NodePort)

	go ntwk.RequestLoop(ntchan, peer.Req_chan, peer.Rep_chan)
	return nil
}

func setupAllPeers(config Configuration) {

	for _, peerAddress := range config.PeerAddresses {
		log.Println("setup  peer ", peerAddress)
		p := ntwk.CreatePeer(peerAddress, config.NodePort)

		err := setupPeerClient(p)
		if err != nil {
			log.Println("connect failed")
			continue
		} else {
			ntwk.MakePingOld(p)
		}
	}

}

//run client against multiple nodes
func runPeermode(option string, config Configuration) {
	log.Println("runPeermode")

	log.Println("setup peers")

	for _, peerAddress := range config.PeerAddresses {

		p := ntwk.CreatePeer(peerAddress, config.NodePort)
		log.Println("add peer ", p)

		err := setupPeerClient(p)
		if err != nil {
			//remove peer
			log.Println("dont add peer to list")
		} else {
			addPeerOut(p)
		}
	}

	switch option {

	case "pingall":

		defer track(runningtime("execute ping"))
		successCount := 0
		for _, peerAddress := range config.PeerAddresses {
			log.Println("setup  peer ", peerAddress, config.NodePort)
			p := ntwk.CreatePeer(peerAddress, config.NodePort)

			err := setupPeerClient(p)
			if err != nil {
				log.Println("connect failed")
				continue
			} else {
				success := ntwk.MakePingOld(p)
				if success {
					successCount++
				}
			}
		}

		log.Println("pinged peers ", len(config.PeerAddresses), " successCount:", successCount)

	case "blockheight":

		for _, peerAddress := range config.PeerAddresses {
			log.Println("setup  peer ", peerAddress)
			p := ntwk.CreatePeer(peerAddress, config.NodePort)

			err := setupPeerClient(p)
			if err == nil {
				log.Println("block height ", p)
				Getblockheight(p)
			}
		}

	}

}

//process replies
func ReplyProcessor(ntchan *ntwk.Ntchan, d time.Duration) {
	log.Println("init ReplyProcessor ")
	for {
		//msg := <-ntchan.Reader_queue

		//reply_string := MsgString(reply)
		//ntchan.Writer_queue <- reply_string

	}
}

//TODO! move
func ping(ntchan ntcl.Ntchan) {

	req_msg := ntwk.EncodeMsgString(ntwk.REQ, ntwk.CMD_PING, "")
	ntchan.REQ_out <- req_msg
	//ntcl.NtwkWrite(ntchan, reqs)

	time.Sleep(1000 * time.Millisecond)

	x := <-ntchan.REP_in
	log.Println("REP_in", x)

}

//run client against single node, just use first IP address in peers i.e. mainpeer
func runSingleMode(option string, config Configuration) {

	//mainPeerAddress := config.PeerAddresses[0]
	//log.Println("setup main peer ", mainPeerAddress, config.NodePort)
	//mainPeer := ntwk.CreatePeer(mainPeerAddress, config.NodePort)
	//log.Println("client with mainPeer ", mainPeer)
	//setupPeerClient(mainPeer)
	//conn := ntwk.OpenConn(mainPeerAddress + ":" + strconv.Itoa(config.NodePort))
	//ntwk.ChannelPeerNetwork(conn, mainPeer)
	//ntchan := ntwk.ConnNtchan(conn, mainPeerAddress)
	ntchan := ntcl.initClient()
	log.Println("init ", ntchan)

	//TODO need to formalize this
	//go Reqprocessor(ntchan)

	//go Reqoutprocessor(ntchan)

	//ntwk.ReaderWriterConnector(ntchan)
	//go ReplyInProcessor(ntchan)

	switch option {

	case "test":
		testclient(ntchan)
		return

	case "ping":
		log.Println("ping")
		//ntwk.MakePingOld(mainPeer)
		ntchan := initClient()
		ping(ntchan)

		time.Sleep(1 * time.Second)

	case "balance":
		log.Println("balance")
		ntchan := initClient()
		Getbalance(ntchan)

		// case "heartbeat":
		// 	log.Println("heartbeat")

		// 	for {
		// 		go ping(ntchan)
		// 		time.Sleep(1 * time.Second)
		// 	}

		// 	// success := ntwk.MakeHandshake(mainPeer)
		// 	// if success {
		// 	// 	log.Println("start heartbeat")
		// 	// 	ntwk.Hearbeat(mainPeer)
		// 	// }

		// case "getbalance":
		// 	log.Println("getbalance")

		// 	Getbalance(mainPeer)

		// case "blockheight":
		// 	log.Println("blockheight")

		// 	Getblockheight(mainPeer)

		// case "faucet":
		// 	log.Println("faucet")
		// 	//get coins
		// 	//GetFaucet(rw)
		// 	GetFaucet(mainPeer)

		// case "txpool":
		// 	_ = Gettxpool(mainPeer)
		// 	return

		// case "pushtx":
		// 	_ = PushTx(mainPeer)
		// 	return

		// case "randomtx":
		// 	_ = MakeRandomTx(mainPeer)
		// 	return

	}
}

//client that only listens to events
func runListenMode(option string, config Configuration) {
	log.Println("listen")

	//ntchan := initClient()

	// mainPeerAddress := config.PeerAddresses[0]
	// log.Println("setup main peer ", mainPeerAddress)
	// mainPeer := ntwk.CreatePeer(mainPeerAddress, config.NodePort)
	// success := ntwk.MakeHandshake(mainPeer)
	// log.Println(success)
	// log.Println("start heartbeat")
	// if success {
	// 	hTime := 2000 * time.Millisecond

	// 	for _ = range time.Tick(hTime) {
	// 		//log.Println(x)
	// 		ntwk.Hearbeat(mainPeer)
	// 	}

	// }
}

//run client without client or server
func runOffline(option string, config Configuration) {

	switch option {
	case "createkeys":
		createKeys()

	case "readkeys":
		kp := readKeys("keys.txt")
		log.Println(kp)

	case "sign":
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter message to sign: ")
		msg, _ := reader.ReadString('\n')
		msg = strings.Trim(msg, string('\n'))
		fmt.Println(msg)
		kp := readKeys("keys.txt")
		signature := crypto.SignMsgHash(kp, msg)
		log.Println("signature ", signature)

		sighex := hex.EncodeToString(signature.Serialize())
		log.Println("sighex ", sighex)

	case "createtx":
		createtx()

	case "verify":
		reader := bufio.NewReader(os.Stdin)
		log.Print("Enter message to verify: ")
		msg, _ := reader.ReadString('\n')
		msg = strings.Trim(msg, string('\n'))
		log.Println(msg)

		fmt.Print("Enter signature to verify: ")
		msgsig, _ := reader.ReadString('\n')
		msgsig = strings.Trim(msgsig, string('\n'))

		sign := crypto.SignatureFromHex(msgsig)

		fmt.Print("Enter pubkey to verify: ")
		msgpub, _ := reader.ReadString('\n')
		log.Println(msgpub)
		msgpub = strings.Trim(msgpub, string('\n'))

		pubkey := crypto.PubKeyFromHex(msgpub)

		verified := crypto.VerifyMessageSignPub(sign, pubkey, msg)
		log.Println("verified ", verified)

	}
}

//dns functions for later, as we can use txt records to get pubkey
func dnslook() {
	//domain := "test.polygonnode.com"
	domain := "swix.io"

	txtrecords, _ := net.LookupTXT(domain)
	// log.Println(txtrecords)

	// for _, txt := range txtrecords {
	// 	fmt.Println(txt)
	// }

	frec := txtrecords[0]
	log.Println("pubkey ", frec)

	// nameserver, _ := net.LookupNS(domain)
	// for _, ns := range nameserver {
	// 	fmt.Println(ns)
	// }

	// cname, _ := net.LookupCNAME(domain)
	// fmt.Println(cname)

	iprecords, _ := net.LookupIP(domain)
	for _, ip := range iprecords {
		fmt.Println("ip ", ip)
	}
}

func getConfig() Configuration {
	file, _ := os.Open("conf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}
	return config
}

func readOption() string {
	optionPtr := flag.String("option", "", "the command to be performed")
	flag.Parse()
	option := *optionPtr
	log.Println("run client with option:", option)
	return option
}

const node_port = 8888

func testclient(ntchan ntcl.Ntchan) {
	// time.Sleep(200 * time.Millisecond)
	// addr := ":" + strconv.Itoa(node_port)
	// log.Println("dial ", addr)
	// conn, err := net.Dial("tcp", addr)
	// if err != nil {
	// 	log.Println("cant run")
	// 	return
	// }

	// log.Println("connected")
	// ntchan := ntcl.ConnNtchan(conn, "client", addr)

	// go ntcl.ReadLoop(ntchan)
	// go ntcl.ReadProcessor(ntchan)
	// go ntcl.WriteProcessor(ntchan)
	// go ntcl.WriteLoop(ntchan, 300*time.Millisecond)

	//subscribe example
	reqs := "REQ#PING#|"
	ntchan.REQ_out <- reqs
	//ntcl.NtwkWrite(ntchan, reqs)

	time.Sleep(1000 * time.Millisecond)

	x := <-ntchan.REP_in
	log.Println("REP_in", x)

	//defer conn.Close()
	return

}

func testclient_subscribe() {
	time.Sleep(200 * time.Millisecond)
	addr := ":" + strconv.Itoa(node_port)
	log.Println("dial ", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("cant run")
		return
	}

	log.Println("connected")
	ntchan := ntcl.ConnNtchan(conn, "client", addr)

	go ntcl.ReadLoop(ntchan)

	//subscribe example
	//reqs := "REQ#PING#|"
	reqs := "REQ#SUBTO#TIME|"
	log.Println("subscribe")
	ntcl.NtwkWrite(ntchan, reqs)

	//log.Println(clientNt.SrcName)

	go func() {
		for {
			rmsg := <-ntchan.Reader_queue
			log.Println("> ", rmsg)
		}
	}()
	time.Sleep(2000 * time.Millisecond)

	reqs = "REQ#SUBUN#TIME|"
	ntcl.NtwkWrite(ntchan, reqs)

	log.Println("unsubscribe")

	time.Sleep(10000 * time.Millisecond)

	//defer conn.Close()
	return

}

//run client based on options
func main() {

	config := getConfig()

	option := readOption()

	//dnslook()

	switch option {

	case "test", "ping", "heartbeat", "balance", "faucet", "txpool", "pushtx", "randomtx":
		runSingleMode(option, config)

	case "createkeys", "sign", "createtx", "verify":
		runOffline(option, config)

	// case "pingall", "blockheight":
	// 	runPeermode(option, config)

	// case "listen":
	// 	runListenMode(option, config)

	default:
		log.Println("unknown option")
	}

}
