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

	peer "github.com/libp2p/go-libp2p-peer"
	//pstore "github.com/libp2p/go-libp2p-peerstore"
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
	peerId peer.ID
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
	peerptr := flag.String("peer", "QmUsWKBMNckswS4gzTCYhTbwkv2cUBPSxShL6pEfnofnKN", "host to connect to")
	libp2pString := "/ip4/127.0.0.1/tcp/8012/ipfs/"
	flag.Parse()

	lc.remote = *hostptr
	lc.port = uint16(*portptr)
	lc.litHomeDir = *dirptr
	lc.peerId = libp2p.GetPeerId(libp2pString + *peerptr)
	log.Println("id of host you'd like to connect to", lc.peerId)
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

	ha, err := libp2p.MakeBasicHost(lc.port, 721) // get randomness
	// set this to be the same as listener for testing
	if err != nil {
		log.Fatal(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	//targetPeerAddr, _ := ma.NewMultiaddr(
	//	fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(lc.peerId)))
	//targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr so we add it to the peerstore
	// so LibP2P knows how to contact it
	//ha.Peerstore().AddAddr(lc.peerId, targetAddr, pstore.PermanentAddrTTL)

	//log.Println("opening stream")
	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set above because
	// we use the same /p2p/1.0.0 protocol
	// s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0") // dial attempt will fail because we're trying to ocnnect two libp2p inst ances which a re on teh same hsots
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// Create a buffered stream so that read and writes are non blocking.
	// bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	// wsConn, err := websocket.Dial(urlString, "", origin)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer wsConn.Close()
	// connect via rpc here
	lc.rpccon = libp2prpc.NewClient(ha, "rpc") // won't work for some reason
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
