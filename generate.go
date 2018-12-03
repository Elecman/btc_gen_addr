package main

import (
	//"crypto/rand"
	"fmt"
	"os"
	"math/big"
	"flag"
	"encoding/hex"
	"strings"
	"time"


     "github.com/btcsuite/btcutil"
     "github.com/btcsuite/btcd/btcec"
     "github.com/btcsuite/btcd/chaincfg"
)



func main() {
	beginTime := time.Now()

	var prefix string
	flag.StringVar(&prefix, "prefix", "1CA6", "prefix you want for your vanity address")
	flag.Parse()

	fmt.Printf("Searching for prefix \"%s\"\n", prefix)
	
	// Initialise big numbers with small numbers
	count, one := big.NewInt(0), big.NewInt(1)

	// Create a slice to pad our count to 32 bytes
	padded := make([]byte, 32)
	var numFound int
	
	for {
		// Increment our counter
		count.Add(count, one)

		// Copy count value's bytes to padded slice
		copy(padded[32-len(count.Bytes()):], count.Bytes())
		//str := copy(padded[32-len(count.Bytes()):], count.Bytes())
		
		_, public := btcec.PrivKeyFromBytes(btcec.S256(), padded)
		// Encode the address and check the prefix.
		caddr, _ := btcutil.NewAddressPubKey(public.SerializeCompressed(), &chaincfg.MainNetParams)
		uaddr, _ := btcutil.NewAddressPubKey(public.SerializeUncompressed(), &chaincfg.MainNetParams)
		//fmt.Printf("%x\n", padded)
		
		if strings.HasPrefix(uaddr.EncodeAddress(), prefix) || strings.HasPrefix(uaddr.EncodeAddress(), prefix) {
			
			//fmt.Printf ("%s\n%s\n%x\n", uaddr.EncodeAddress(), caddr.EncodeAddress(), padded)
			
			numFound++
			fmt.Printf("\nElapsed: %s\nUaddr: %s\nCaddr: %s\nwif: %x\nnumfound: %d\n",
				time.Since(beginTime), uaddr.EncodeAddress(), caddr.EncodeAddress(), padded,
				numFound)

			file, err := os.OpenFile("for_zerro.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("Error pushing data to file: %s", err)
				os.Exit(666)
			}
			if _, err := file.WriteString(uaddr.EncodeAddress() + " " + caddr.EncodeAddress() + " " + hex.EncodeToString(padded) + "\n"); err != nil {
				fmt.Printf("Error pushing data to file: %s", err)
				os.Exit(666)
			}
		}
	}
}

func init() {
	// Panic on init if the assumptions used by the code change.
	if btcec.PubKeyBytesLenUncompressed != 65 {
		panic("Source code assumes 65-byte uncompressed secp256k1 " +
			"serialized public keys")
	}
}

