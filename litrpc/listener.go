package litrpc

import (
	"log"
	"net/http"
	"net/rpc"
	"strconv"

	"github.com/mit-dci/lit/qln"
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

func RPCListen(rpcl *LitRPC, port uint16) {

	rpc.Register(rpcl)
	localport := ":" + strconv.Itoa(int(port))
	log.Println("Serving conn through tls on port", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	server := &http.Server{
		Addr:         localport,
		Handler:      mux,
	}
	log.Fatal(server.ListenAndServeTLS("server.crt", "server.key"))
}
