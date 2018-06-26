package qln

import (
	"fmt"
	"log"
	"time"

	"github.com/adiabat/bech32"
	"github.com/btcsuite/fastsha256"
	litconfig"github.com/mit-dci/lit/config"
)

// AutoReconnect will start listening for incoming connections
// and attempt to automatically reconnect to all
// previously known peers.
func (nd *LitNode) AutoReconnect(listenPort string, interval int64, config litconfig.Config) {
	// Listen myself after a timeout
	nd.TCPListener(listenPort)

	// Reconnect to other nodes after an interval
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		for {
			fmt.Println("Reconnecting to known peers")
			var empty [33]byte
			i := uint32(1)
			for {
				pubKey, _ := nd.GetPubHostFromPeerIdx(i)
				if pubKey == empty {
					log.Printf("Done, tried %d hosts\n", i-1)
					break
				}

				nd.RemoteMtx.Lock()
				_, alreadyConnected := nd.RemoteCons[i]
				nd.RemoteMtx.Unlock()

				if alreadyConnected {
					i++
					continue
				}

				idHash := fastsha256.Sum256(pubKey[:])
				adr := bech32.Encode("ln", idHash[:20])

				err := nd.DialPeer(adr, config)

				if err != nil {
					log.Printf("Could not restore connection to %s: %s\n", adr, err.Error())
				}

				i++

			}
			<-ticker.C
		}
	}()

}
