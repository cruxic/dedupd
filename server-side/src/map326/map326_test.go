package map326

import (
	"testing"
	"github.com/stretchr/testify/require"
	"util"
	"math/rand"
	"fixedpool"
)

func uint16ToBytes(val int, dest []byte) {
	//big endian
	dest[0] = byte((val >> 8) & 0xFF)
	dest[1] = byte(val & 0xFF)
}


func Test_Entry(t * testing.T) {
	req := require.New(t)

	e := _Entry(make([]byte, entrySize))

	//all zeros
	req.Equal(fixedpool.Zero, e.getPtr())
	req.True(e.cmpKeySuffix(make([]byte, KeySize-2)) == 0)
	req.Equal(ValueFromInt(0), e.getValue())

	//set ptr, key and value
	e.setPtr(fixedpool.Ptr(0xDeadBeef))
	keySuff := make([]byte, KeySize - 2)
	util.FillSeq(keySuff, 13)
	e.setKeyValue(keySuff, ValueFromInt(0xBabeFace))

	//verify
	req.Equal(fixedpool.Ptr(0xDeadBeef), e.getPtr())
	req.True(e.cmpKeySuffix(keySuff) == 0)
	req.Equal(ValueFromInt(0xBabeFace), e.getValue())

	//set only value
	v := ValueFromInt(0xFaceF00d)
	e.setValue(v)
	req.Equal(v, e.getValue())

	//deeper testing of cmpKeySuffix
	for i := 0; i < len(keySuff); i++ {
		req.True(e.cmpKeySuffix(keySuff) == 0)
		keySuff[i]++
		req.False(e.cmpKeySuffix(keySuff) == 0)
		keySuff[i]--
	}
}

func TestInternals(t * testing.T) {
	req := require.New(t)

	req.Equal(0x1234, uint16FromBytes([]byte{0x12, 0x34}))
}

func Test_Basic(t * testing.T) {
	req := require.New(t)

	dm, err := New(99)  //will be rounded up to 1 entry per region
	req.Nil(err)
	req.NotNil(dm)

	//Create 5 keys having the same first 31 bytes
	keys := make(map[[KeySize]byte]Value)
	var k [KeySize]byte
	util.FillSeq(k[:], 1)
	k[KeySize-1] = 0
	for i := 0; i < 5; i++ {
		keys[k] = ValueFromInt(1234567890 * i)
		k[KeySize-1]++
	}

	//Put all 5
	for k, val := range keys {
		req.Equal(1, dm.Put(k[:], val))

		//verify
		val2, found := dm.Get(k[:])
		req.True(found)
		req.Equal(val, val2)
	}
}

func randKeyValue() (key [KeySize]byte, value Value) {
	rand.Read(key[:])
	rand.Read(value[:])
	return
}

func Test_randFill(t * testing.T) {
	req := require.New(t)

	approxNumKeys := nRegions * 16
	dm, err := New(approxNumKeys)
	req.Nil(err)
	req.NotNil(dm)
	req.Equal(4, dm.epr)

	rand.Seed(99)

	//Add random keys until full
	nAdded := 0
	for {
		k, v := randKeyValue()
		if dm.Put(k[:], v) == 0 {
			break
		}

		nAdded++
		if nAdded > 2 * approxNumKeys {
			t.Error("too many iterations")
			return
		}
	}

	//Added at least 99.0% of approxNumKeys
	percent := float32(nAdded) / float32(approxNumKeys) * 100.0
	req.True(percent > 99.0, percent)

	//
	// Verify all

	rand.Seed(99)

	for i := 0; i < nAdded; i++ {
		k, v := randKeyValue()
		v2, found := dm.Get(k[:])
		if !found {
			t.Errorf("key %d not found!", i)
			return
		} else if v != v2 {
			t.Errorf("wrong value for key %d", i)
			return
		}
	}
}

func Test_keyNotFound(t * testing.T) {
	req := require.New(t)

	dm, err := New(nRegions * 8)
	req.Nil(err)

	//not found when empty
	k, v := randKeyValue()
	_, found := dm.Get(k[:])
	req.False(found)

	//add one key
	req.True(dm.Put(k[:], v) != 0)

	//find it
	_, found = dm.Get(k[:])
	req.True(found)

	//tweak key so it's in the same bucket but not found
	k[KeySize - 2]++
	_, found = dm.Get(k[:])
	req.False(found)

	//add it and find it
	req.True(dm.Put(k[:], v) != 0)
	_, found = dm.Get(k[:])
	req.True(found)

	//tweak again
	k[KeySize - 1]++
	_, found = dm.Get(k[:])
	req.False(found)
}


func benchRandFill(approxNumKeys int, keys []KV) int {
	dm, _ := New(approxNumKeys)

	//Add random keys until full
	nAdded := 0
	for _, kv := range keys {
		if dm.Put(kv.K[:], kv.V) == 0 {
			break
		}

		nAdded++
	}

	return nAdded
}

var noCompilerOptimize int

func Benchmark_randFill(b *testing.B) {
	//make some random keys
	approxNumKeys := nRegions * 20

	keys := make([]KV, 0, approxNumKeys)
	var kv KV
	for i := 0; i < approxNumKeys; i++{
		kv.K, kv.V = randKeyValue()
		keys = append(keys, kv)
	}

	b.Run("randFill", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			noCompilerOptimize += benchRandFill(approxNumKeys, keys)
		}
	})
}

type KV struct {
	K [KeySize]byte
	V Value
}

func benchRandRead(dm *Map, keys []KV) bool {
	//call Map.Get() on each key and verify correct value
	for _, kv := range keys {
		v, found := dm.Get(kv.K[:])
		if !found {
			//should not happen
			return false
		} else if v != kv.V {
			//should not happen
			return false

		}
	}

	return true
}


/*
Benchmark random access of every key in a full map.
*/
func Benchmark_randRead(b *testing.B) {
	approxNumKeys := 16000000 //nRegions * 40
	dm, _ := New(approxNumKeys)

	rand.Seed(1234)

	keys := make([]KV, 0, approxNumKeys)
	var kv KV

	//Add random keys until full
	for {
		kv.K, kv.V = randKeyValue()
		if dm.Put(kv.K[:], kv.V) == 0 {
			break
		}

		keys = append(keys, kv)
	}

	swapFunc := func(i, j int) {
		temp := keys[i]
		keys[i] = keys[j]
		keys[j] = temp
	}

	rand.Shuffle(len(keys), swapFunc)

	b.Run("randRead", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			//shuffle the keys


			if !benchRandRead(dm, keys) {
				panic("benchRandRead error")
			}
		}
	})

	//fmt.Println("nEarly", nEarly, "pointless", nPointless)
}

/*
func Benchmark_cmp1(b *testing.B) {
	rand.Seed(13)
	k1, _ := randKeyValue()

	for j := 0; j < b.N; j++ {
		k2, _ := randKeyValue()
		if bytes.Equal(k1[:], k2[:]) {
			panic("equal")
		}
	}
}

func Benchmark_cmp2(b *testing.B) {
	rand.Seed(13)
	k1, _ := randKeyValue()

	for j := 0; j < b.N; j++ {
		k2, _ := randKeyValue()
		if bytes.Compare(k1[:], k2[:]) == 0 {
			panic("equal")
		}
	}
}
*/

/*
type K [KeySize]byte

func benchRandFillGoMap() int {
	approxNumKeys := nRegions * 40

	m := make(map[K]Value, approxNumKeys)

	rand.Seed(99)

	for i := 0; i < approxNumKeys; i++ {
		k, v := randKeyValue()
		m[k] = v
	}

	return len(m)
}

func Benchmark_GoMap(b *testing.B) {
	for j := 0; j < b.N; j++ {
		noCompilerOptimize += benchRandFillGoMap()
	}
}
*/
