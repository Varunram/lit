package main

import (
	"bytes"
	"fmt"

	"github.com/adiabat/btcd/chaincfg/chainhash"
	"github.com/mit-dci/lit/lnutil"
)

func main() {
	// given an uint64, split into 2 bytes.
	// byte1: 5 bit shift, 2 bit MSBs
	// byte2: 7 bit LSBs

	/*
		for i := uint64(1000); i < 1100; i++ {
			ba, bb, err := split(i)
			if err != nil {
				panic(err)
			}
			//		fmt.Printf("ba:%x bb:%x\n", ba, bb)

			r := join(ba, bb)
			fmt.Printf("rejoin %d\n", r)

		}
	*/

	privRoot := chainhash.HashH([]byte("my private key"))

	ka, _ := deriveK(privRoot, "asset 1 a")

	//	fmt.Printf("key: %x\n", privRoot[:])

	msg := chainhash.HashH([]byte("hi"))

	var hits, tries int

	tries = 10000

	for i := 0; i < tries; i++ {

		ka = chainhash.HashH(ka[:])
		msg = chainhash.HashH(msg[:])
		privRoot = chainhash.HashH(privRoot[:])

		if oneTry(privRoot, ka, msg) {
			hits++
		}
	}

	fmt.Printf("%d tries %d hits\n", tries, hits)

	return
}

func oneTry(a, k, m [32]byte) bool {
	pubRootArr := lnutil.PubFromHash(a)

	R := lnutil.PubFromHash(k)

	s, err := RSign(m, a, k)
	if err != nil {
		panic(err)
	}
	fmt.Printf("s: %x\n", s)
	sGarr := lnutil.PubFromHash(s)

	sGpred, err := SGpredict(pubRootArr, R, m)
	if err != nil {
		panic(err)
	}

	if bytes.Equal(sGarr[:], sGpred.SerializeCompressed()) {
		return true
	}

	return false
}
