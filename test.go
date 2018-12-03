package main

import (
	"log"
	"math/rand"
	"encoding/hex"
	"time"
	"sync"
	"runtime"
	"os"
	"fmt"
	//"flag"
	//"strings"
	
	"io"
	"encoding/csv"
	"strconv"
	"bufio"
	
     "github.com/btcsuite/btcutil"
     "github.com/btcsuite/btcd/btcec"
     "github.com/btcsuite/btcd/chaincfg"
)

type concurrentMap struct {
	sync.Mutex
	addresses map[string]bool
}

var partitions = int(6)
var count int64
var oldCount int64
var beginTime = time.Now()
var numFound int
var prefix string
var addressesMap = concurrentMap { addresses: make(map[string]bool), }
	
func main() {
	//flag.StringVar(&prefix, "prefix", "1CA6", "prefix you want for your vanity address")
	//flag.Parse()
	//fmt.Printf("Searching for prefix \"%s\"\n", prefix)
	runtime.GOMAXPROCS(runtime.NumCPU() + 1)

	count = int64(0)
	
	
	loadAddresses()

	value, _ := time.ParseDuration("1s")
	checkTimer := time.NewTimer(value)
	go func() {
		for {
			select {
			case <-checkTimer.C:
				log.Printf("Checked: %d, Speed: %d per second", count, count-oldCount)
				oldCount = count
				checkTimer.Reset(value)
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < partitions; i++ {
		wg.Add(1)
		addr := generateSeedAddress()
		log.Printf("Seed addr: %x\n", addr)
		go generateAddresses(addr)
	}
	wg.Wait()
}

func loadAddresses() int64 {
	processedBlocks := int64(0)
	count := int64(0)
	f, _ := os.Open("./balances.csv")
	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	first := true
	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}

		if first {
			processedBlocks, err = strconv.ParseInt(record[0], 10, 64)
			if err != nil {
				log.Panic(err)
			}
			first = !first
			continue
		}

		count++

		addressesMap.Lock()
		addressesMap.addresses[record[0]] = true
		addressesMap.Unlock()
	}
	log.Printf("Number of addresses loaded: %d", count)

	return processedBlocks
}

func generateSeedAddress() []byte {
	
	paddedR := make([]byte, 32)
	rand.Read(paddedR)
	return paddedR
}

func generateAddresses(paddedR []byte) {
	
	for ; ; {
				
		incrementPrivKey(paddedR)
		
		_, public := btcec.PrivKeyFromBytes(btcec.S256(), paddedR)
		uaddr, _ := btcutil.NewAddressPubKey(public.SerializeUncompressed(), &chaincfg.MainNetParams)
		caddr, _ := btcutil.NewAddressPubKey(public.SerializeCompressed(), &chaincfg.MainNetParams)
		addressesMap.Lock()
		if _, ok := addressesMap.addresses[uaddr.EncodeAddress()]; ok {
			log.Printf("priv: %x, addr: %s", paddedR, uaddr.EncodeAddress())
			writeToFound(fmt.Sprintf("Private: %s, Address: %s\n", hex.EncodeToString(paddedR), uaddr.EncodeAddress()))
		}
		
		if _, ok := addressesMap.addresses[caddr.EncodeAddress()]; ok {
			log.Printf("priv: %x, addr: %s", paddedR, caddr.EncodeAddress())
			writeToFound(fmt.Sprintf("Private: %s, Address: %s\n", hex.EncodeToString(paddedR), caddr.EncodeAddress()))
		}
		addressesMap.Unlock()
		//if strings.HasPrefix(uaddr.EncodeAddress(), prefix) || strings.HasPrefix(caddr.EncodeAddress(), prefix) {
			
		//	numFound++
		//	fmt.Printf("\nElapsed: %s\nUaddr: %s\nCaddr: %s\nwif: %x\nnumfound: %d\n",
		//		time.Since(beginTime), uaddr.EncodeAddress(), caddr.EncodeAddress(), paddedR,
		//		numFound)
				
		//	file, err := os.OpenFile("for_zerro.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		//	if err != nil {
		//		fmt.Printf("Error pushing data to file: %s", err)
		//		os.Exit(666)
		//	}
		//	if _, err := file.WriteString(uaddr.EncodeAddress() + " " + caddr.EncodeAddress() + " " + hex.EncodeToString(paddedR) + "\n"); err != nil {
		//		fmt.Printf("Error pushing data to file: %s", err)
		//		os.Exit(666)
		//	}
		//}
		count++
	}
}

func writeToFound(text string) {
	foundFileName := "./found.txt"
	if _, err := os.Stat(foundFileName); os.IsNotExist(err) {
		_, _ = os.Create(foundFileName)
	}
	f, err := os.OpenFile(foundFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	defer f.Close()
	if err != nil {
		log.Printf(err.Error())
	}
	_, err = f.WriteString(text)
	if err != nil {
		log.Printf(err.Error())
	}
}

func incrementPrivKey(paddedR []byte) {
	for i := 31; i > 0; i-- {
		if paddedR[i]+1 == 255 {
			paddedR[i] = 0
		} else {
			paddedR[i] += 1
			break
		}
	}
}
