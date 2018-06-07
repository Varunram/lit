package litrpc

import (
	"bytes"
	//"fmt"
	"io/ioutil"
	"log"
	//"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"golang.org/x/net/websocket"
	"github.com/mit-dci/lit/libp2p"
	"github.com/mit-dci/lit/qln"
	libp2prpc "github.com/hsanjuan/go-libp2p-gorpc"
)

/*
Remote Procedure Calls
RPCs are how people tell the lit node what to do.
It ends up being the root of ~everything in the executable.

*/

// A LitRPC is the user I/O interface; it owns and initialized a SPVCon and LitNode
// and listens and responds on RPC

type LitRPC struct {
	Node      *qln.LitNode
	OffButton chan bool
}

func serveWS(ws *websocket.Conn) {
	body, err := ioutil.ReadAll(ws.Request().Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return
	}

	log.Printf(string(body))
	ws.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))

	jsonrpc.ServeConn(ws)
}

func RPCListen(rpcl *LitRPC, host string, port uint16) {

	rpc.Register(rpcl)

	//listenString := fmt.Sprintf("%s:%d", host, port)

	ha, err := libp2p.MakeBasicHost(port, 722) // get a random source for this
	if err != nil {
		log.Fatal(err)
	}
	log.Println("The port is:", port)
	log.Println("ID: ", ha.ID())

	server := libp2prpc.NewServer(ha, "rpc")
	libp2prpc.NewClientWithServer(ha, "rpc", server)
	//server := libp2prpc.NewServer(ha, "tcp")
	//serverClient := libp2prpc.NewClientWithServer(ha, "tcp", server)
	//rpcServer := libp2prpc.NewServer(ha, "tcp")
	//http.Handle("/ws", websocket.Handler(serveWS))
	//log.Fatal(http.ListenAndServe(listenString, nil))
}
