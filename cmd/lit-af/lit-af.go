package main

import (
	"flag"
	"fmt"
	//"bufio"
	//"context"
	"log"
	//"net/rpc"
	//"net/rpc/jsonrpc"
	"os"
	"path/filepath"
	"strings"

	//"golang.org/x/net/websocket"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/lit/libp2p"

	ma "github.com/multiformats/go-multiaddr"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	libp2prpc "github.com/hsanjuan/go-libp2p-gorpc"
)

/*
Lit-AF

The Lit Advanced Functionality interface.
This is a text mode interface to lit.  It connects over jsonrpc to the a lit
node and tells that lit node what to do.  The lit node also responds so that
lit-af can tell what's going on.

lit-gtk does most of the same things with a gtk interface, but there will be
some yet-undefined advanced functionality only available in lit-af.

May end up using termbox-go

*/

//// BalReply is the reply when the user asks about their balance.
//// This is a Non-Channel
//type BalReply struct {
//	ChanTotal         int64
//	TxoTotal          int64
//	SpendableNow      int64
//	SpendableNowWitty int64
//}

const (
	litHomeDirName  = ".lit"
	historyFilename = "lit-af.history"
)

type litAfClient struct {
	remote string
	port   uint16
	rpccon *libp2prpc.Client
	//httpcon
	litHomeDir string
}

type Command struct {
	Format           string
	Description      string
	ShortDescription string
}

func setConfig(lc *litAfClient) {
	hostptr := flag.String("node", "127.0.0.1", "host to connect to")
	portptr := flag.Int("p", 8001, "port to connect to")
	dirptr := flag.String("dir", filepath.Join(os.Getenv("HOME"), litHomeDirName), "directory to save settings")

	flag.Parse()

	lc.remote = *hostptr
	lc.port = uint16(*portptr)
	lc.litHomeDir = *dirptr
}

// for now just testing how to connect and get messages back and forth
func main() {
	lc := new(litAfClient)
	setConfig(lc)

	//	dialString := fmt.Sprintf("%s:%d", lc.remote, lc.port)

	/*
		client, err := net.Dial("tcp", dialString)
		if err != nil {
			log.Fatal("dialing:", err)
		}
		defer client.Close()
	*/

	//	dialString := fmt.Sprintf("%s:%d", lc.remote, lc.port)
	// origin := "http://127.0.0.1/"
	// urlString := fmt.Sprintf("ws://%s:%d/ws", lc.remote, lc.port)
	//	url := "ws://127.0.0.1:8000/ws"

	ha, err := libp2p.MakeBasicHost(lc.port, 0)
	// don't pass a random seed and don't ask the user to provide this
	if err != nil {
		log.Fatal(err)
	}

	ha.SetStreamHandler("/p2p/1.0.0", libp2p.HandleStream)

	// The following code extracts target's peer ID from the
	// given multiaddress

	// force the user to input this, no other way?
	p2pstring := "/ip4/127.0.0.1/tcp/8001/ipfs/"
	p2pPeer := "QmaBwhATQoBinrwfi8cU2wNY8SYbJUULCSC3Y4dhSGb2ce"
	ipfsaddr, err := ma.NewMultiaddr(p2pstring + p2pPeer)
	if err != nil {
		log.Fatalln(err)
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Fatalln(err)
	}

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr so we add it to the peerstore
	// so LibP2P knows how to contact it
	ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

	log.Println("opening stream")
	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set above because
	// we use the same /p2p/1.0.0 protocol
	// s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// Create a buffered stream so that read and writes are non blocking.
	//rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// wsConn, err := websocket.Dial(urlString, "", origin)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer wsConn.Close()

	lc.rpccon = libp2prpc.NewClientWithServer(ha, "tcp", libp2prpc.NewServer(ha, "tcp"))

	go lc.RequestAsync()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       lnutil.Prompt("lit-af") + lnutil.White("# "),
		HistoryFile:  filepath.Join(lc.litHomeDir, historyFilename),
		AutoComplete: lc.NewAutoCompleter(),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

	// main shell loop
	for {
		// setup reader with max 4K input chars
		msg, err := rl.Readline()
		if err != nil {
			break
		}
		msg = strings.TrimSpace(msg)
		if len(msg) == 0 {
			continue
		}
		rl.SaveHistory(msg)

		cmdslice := strings.Fields(msg)                         // chop input up on whitespace
		fmt.Fprintf(color.Output, "entered command: %s\n", msg) // immediate feedback

		err = lc.Shellparse(cmdslice)
		if err != nil { // only error should be user exit
			log.Fatal(err)
		}
	}
}
