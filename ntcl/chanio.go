package ntcl

// functions relating to network stack

// network layer (NTCL)

// NTCL -> semantics of channels
// TCP/IP -> golang net

// golang has the native net package. as TCP only deals with byte streams we need some form
// to delinate distinct messages to implement the equivalent of actors. we have
// channels as the major building block for the networkw

// the struct ntchan is the struct wraps the native readwriter with reads and write queues
// as channels.

// network reads happen in distinct units of messages which are delimited by the DELIM byte
// messages have types to indicate the flow of message direction and commands
// an open question is use of priorities, timing etc.

// the P2P network or any network connection has different behaviour based on the
// types of messages going through it. a request-reply for example will have a single read
// and single write in order, publish-subscribe will  push messages from producers to
// consumers, etc.

// we have only one single two-way channel available as we are on a single
// socket we need to coordinate the reads and writes. the network is a scarce
// resource and depending on the context and semantics messages will be sent/received in
// different style. golangs channels don't fully map to underlying network semantics.
// TCP does not have waiting block or buffering

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/polygonledger/node/ntwk"
)

const (
	EMPTY_MSG  = "EMPTY"
	ERROR_READ = "error_read"
)

//network channel
type Ntchan struct {
	//TODO is only single connection
	Conn     net.Conn
	SrcName  string
	DestName string
	//TODO message type
	Reader_queue chan string
	Writer_queue chan string
	//inflow
	REQ_in  chan string
	REP_out chan string
	//outflow
	REQ_out chan string
	REP_in  chan string

	PUB_time_out  chan string
	SUB_time_out  chan string
	PUB_time_quit chan int
	// SUB_request_out   chan string
	// SUB_request_in    chan string
	// UNSUB_request_out chan string
	// UNSUB_request_in  chan string

	// Reader_processed int
	// Writer_processed int
}

func vlog(s string) {
	verbose := true
	if verbose {
		log.Println(s)
	}
}

func logmsgd(src string, msg string) {
	s := fmt.Sprintf("[%s] ### %v", src, msg)
	vlog(s)
}

func logmsgc(name string, src string, msg string) {
	s := fmt.Sprintf("%s [%s] ### %v", name, src, msg)
	vlog(s)
}

//wrap connection in Ntchan
func ConnNtchan(conn net.Conn, SrcName string, DestName string) Ntchan {
	var ntchan Ntchan
	ntchan.Reader_queue = make(chan string)
	ntchan.Writer_queue = make(chan string)
	ntchan.REQ_in = make(chan string)
	ntchan.REP_in = make(chan string)
	ntchan.REP_out = make(chan string)
	ntchan.REQ_out = make(chan string)
	ntchan.PUB_time_out = make(chan string)
	ntchan.PUB_time_quit = make(chan int)
	// ntchan.Reader_processed = 0
	// ntchan.Writer_processed = 0
	ntchan.Conn = conn
	ntchan.SrcName = SrcName
	ntchan.DestName = DestName

	return ntchan
}

//for testing
func ConnNtchanStub(name string) Ntchan {
	var ntchan Ntchan
	ntchan.Reader_queue = make(chan string)
	ntchan.Writer_queue = make(chan string)
	ntchan.REQ_in = make(chan string)
	ntchan.REP_in = make(chan string)
	ntchan.REP_out = make(chan string)
	ntchan.REQ_out = make(chan string)
	ntchan.PUB_time_out = make(chan string)
	ntchan.PUB_time_quit = make(chan int)
	//ntchan.Reader_processed = 0
	//ntchan.Writer_processed = 0

	return ntchan
}

func ReadLoop(ntchan Ntchan) {
	vlog("init ReadLoop " + ntchan.SrcName + " " + ntchan.DestName)
	d := 300 * time.Millisecond
	//msg_reader_total := 0
	for {
		//read from network and put in channel
		vlog("iter ReadLoop " + ntchan.SrcName + " " + ntchan.DestName)
		msg, err := MsgRead(ntchan)
		if err != nil {

		}
		//handle cases
		//currently can be empty or len, shoudl fix one style
		if len(msg) > 0 { //&& msg != EMPTY_MSG {
			vlog("ntwk read => " + msg)
			logmsgc(ntchan.SrcName, "ReadLoop", msg)
			vlog("put " + msg)
			//put in the queue to process
			ntchan.Reader_queue <- msg
		}

		time.Sleep(d)
		//fix: need ntchan to be a pointer
		//msg_reader_total++
	}
}

//read from reader queue and process by forwarding to right channel
func ReadProcessor(ntchan Ntchan) {

	for {
		msgString := <-ntchan.Reader_queue
		logmsgd("ReadProcessor", msgString)

		if len(msgString) > 0 {
			logmsgc(ntchan.SrcName, "ReadProcessor", msgString) //, ntchan.Reader_processed)
			//TODO try catch

			msg := ntwk.ParseMessage(msgString)

			if msg.MessageType == ntwk.REQ {

				msg_string := ntwk.MsgString(msg)
				logmsgd("ReadProcessor", "REQ_in")

				//TODO!
				ntchan.REQ_in <- msg_string
				// reply_string := "echo:" + msg_string
				// log.Println(">> ", reply_string)
				// ntchan.Writer_queue <- reply_string

			} else if msg.MessageType == ntwk.REP {
				//TODO!
				//msg_string := MsgString(msg)
				msg_string := ntwk.MsgString(msg)
				logmsgd("ReadProcessor", "REP_in")
				ntchan.REP_in <- msg_string

				x := <-ntchan.REP_in
				vlog("x " + x)
			}

			//ntchan.Reader_processed++
			//log.Println(" ", ntchan.Reader_processed, ntchan)
		}
	}

}

//process from higher level chans into write queue
func WriteProcessor(ntchan Ntchan, d time.Duration) {
	for {
		msg_string := <-ntchan.REP_out
		//TODO! select reqout
		//msg_string := <-ntchan.REQ_out
		log.Println("writeprocessor ", msg_string)
		ntchan.Writer_queue <- msg_string
		time.Sleep(d)
	}
}

func WriteLoop(ntchan Ntchan, d time.Duration) {
	//msg_writer_total := 0
	for {
		//log.Println("loop writer")
		//TODO!
		//

		//take from channel and write
		msg := <-ntchan.Writer_queue
		vlog("writeloop " + msg)
		NtwkWrite(ntchan, msg)
		//logmsg(ntchan.Name, "WriteLoop", msg, msg_writer_total)
		//NetworkWrite(ntchan, msg)

		time.Sleep(d)
		//msg_writer_total++
	}
}

func PublishTime(ntchan Ntchan) {
	timeFormat := "2006-01-02T15:04:05"
	limiter := time.Tick(1000 * time.Millisecond)
	pubcount := 0
	log.Println("PublishTime")

	for {
		t := time.Now()
		tf := t.Format(timeFormat)
		vlog("pub " + tf)
		ntchan.PUB_time_out <- tf
		<-limiter
		pubcount++
	}

}

//publication to writer queue. requires quit channel
func PubWriterLoop(ntchan Ntchan) {

	for {
		select {
		case msg := <-ntchan.PUB_time_out:
			vlog("sub " + msg)
			ntchan.Writer_queue <- msg
		case <-ntchan.PUB_time_quit:
			fmt.Println("stop pub")
			return
			// default:
			// 	fmt.Println("no message received")
		}
		time.Sleep(50 * time.Millisecond)

	}

}
