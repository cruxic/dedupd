package main

import (
	"math/rand"
	"time"
	"log"
	"fmt"
	"map326"
)

type KV struct {
	K [map326.KeySize]byte
	V map326.Value
}

func randKeyValue() (kv KV) {
	rand.Read(kv.K[:])
	rand.Read(kv.V[:])
	return
}



/*
Benchmark random access of every key in a full map.
*/
func benchRandRead() {
	t := time.Now()
	approxNumKeys := 16000000 //nRegions * 40
	dm, _ := map326.New(approxNumKeys)
	fmt.Printf("alloc took %s\n", time.Since(t))

	rand.Seed(1234)

	keys := make([]KV, approxNumKeys)
	t = time.Now()
	for i := 0; i < approxNumKeys; i++ {
		keys[i] = randKeyValue()
	}
	fmt.Printf("gen keys took %s\n", time.Since(t))

	//Add random keys until full
	t = time.Now()
	nKeys := 0
	for _, kv := range keys {
		if dm.Put(kv.K[:], kv.V) == 0 {
			break
		}
		nKeys++
	}

	fmt.Printf("fill %d keys took %s\n", nKeys, time.Since(t))

	//truncate
	keys = keys[0:nKeys]

	swapFunc := func(i, j int) {
		temp := keys[i]
		keys[i] = keys[j]
		keys[j] = temp
	}

	t = time.Now()
	rand.Shuffle(len(keys), swapFunc)
	fmt.Printf("shuffle took %s\n", time.Since(t))

	t = time.Now()
	for _, kv := range keys {
		v, found := dm.Get(kv.K[:])
		if !found {
			//should not happen
			log.Fatal("not found")
		} else if v != kv.V {
			//should not happen
			log.Fatal("wrong value")

		}
	}
	fmt.Printf("read took %s\n", time.Since(t))
}

func main() {
	benchRandRead()
}
