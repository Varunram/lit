package libp2p

import (
	"bufio"
	"context"
	"crypto/rand"
	// "encoding/json"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"reflect"
	"sync"
	"time"
	// "bytes"
	"os"
	"strings"
	"strconv"

	//golog "github.com/ipfs/go-log"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p/gxlibs/github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p/gxlibs/github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p/gxlibs/github.com/libp2p/go-libp2p-net"
	ma "github.com/libp2p/go-libp2p/gxlibs/github.com/multiformats/go-multiaddr"
)

var mutex = &sync.Mutex{}
// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. Use secio.
func MakeBasicHost(listenPort uint16, randseed int64) (host.Host, error) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		// equivalent of listenString in listener.go
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("serving libp2p connection on %s\n", fullAddr)
	return basicHost, nil
}

func HandleStream(s net.Stream) {

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	log.Println(reflect.TypeOf(rw))

	go readData(rw)
	go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {

			mutex.Lock()
			// do your stuff here, don't bother too much
			mutex.Unlock()
		}
	}
}

func writeData(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			// do your stuff
			mutex.Unlock()
			mutex.Lock()
			//rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		_, err = strconv.Atoi(sendData)
		// bm, err := srtconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}

		mutex.Lock()
		rw.Flush()
		// do your stuff here
		mutex.Unlock()
	}

}
