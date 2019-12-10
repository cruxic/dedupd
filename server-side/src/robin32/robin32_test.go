package robin32

import (
	"testing"
	"github.com/stretchr/testify/require"
	"util"
	"math/rand"
	//"fixedpool"
	"encoding/binary"
	"bytes"
)

func randKey() []byte {
	k := make([]byte, KeySize)
	rand.Read(k)  //never fails
	return k
}

func randValue(valSize int) []byte {
	v := make([]byte, valSize)
	rand.Read(v)  //never fails
	return v
}

func TestInternals(t * testing.T) {
	req := require.New(t)

	m := NewMap(100, 13)
	poolSize := m.pool.NumFree()

	//
	// allocBucket
	k1 := util.MakeSeq(KeySize, 1)
	v1 := util.MakeSeq(13, 3)
	b1, ok := m.allocBucket(k1, v1)
	req.True(ok)
	req.False(b1.isEmpty())

	//
	// keyPrefixAsUint32
	req.Equal(uint32(0x04030201), keyPrefixAsUint32(k1))


	//
	// isKeyEqual2
	req.True(m.isKeyEqual2(b1, b1.KeyPrefix, k1[4:]))
	req.False(m.isKeyEqual2(b1, b1.KeyPrefix + 1, k1[4:]))
	k1[4]++
	req.False(m.isKeyEqual2(b1, b1.KeyPrefix, k1[4:]))
	k1[4]--
	k1[KeySize-1]++
	req.False(m.isKeyEqual2(b1, b1.KeyPrefix, k1[4:]))
	k1[KeySize-1]--
	req.True(m.isKeyEqual2(b1, b1.KeyPrefix, k1[4:]))

	//
	// isKeyEqual
	req.True(m.isKeyEqual(b1, b1))
	k1[KeySize-1]++
	b2, ok := m.allocBucket(k1, v1)
	k1[KeySize-1]--
	req.True(ok)
	req.False(m.isKeyEqual(b1, b2))
	m.freeBucket(&b2)
	b2, ok = m.allocBucket(k1, v1)
	req.True(ok)
	req.True(m.isKeyEqual(b1, b2))

	//
	// getValueRef
	req.Equal(v1, m.getValueRef(b1))
	req.Equal(v1, m.getValueRef(b2))

	//
	// copyValue
	v3 := util.MakeSeq(13, 0x99)
	b3, ok := m.allocBucket(k1, v3)
	req.True(ok)
	req.Equal(v3, m.getValueRef(b3))
	m.copyValue(b1, b3)
	req.Equal(v1, m.getValueRef(b3))

	//
	// index
	nBuckets := len(m.buckets)
	req.Equal(0, m.index(0))
	req.Equal(13, m.index(13))
	req.Equal(nBuckets - 1, m.index(uint32(nBuckets - 1)))
	req.Equal(0, m.index(uint32(nBuckets)))
	req.Equal(1, m.index(uint32(nBuckets + 1)))

	//
	// probeDist
	var b _Bucket
	b.KeyPrefix = 3
	req.True(b.isEmpty())
	req.Equal(0, m.probeDist(b, 3))
	req.Equal(8, m.probeDist(b, 11))
	req.Equal(nBuckets - 1 - 3, m.probeDist(b, nBuckets - 1))
	req.Equal(nBuckets - 1 - 2, m.probeDist(b, 0))
	req.Equal(nBuckets - 1, m.probeDist(b, 2))

	//
	// freeBucket
	req.False(b1.isEmpty())
	m.freeBucket(&b1)
	req.True(b1.isEmpty())
	m.freeBucket(&b2)
	m.freeBucket(&b3)
	//all memory was returned to pool
	req.Equal(poolSize, m.pool.NumFree())
}

func int2key(i int) []byte {
	var four [4]byte
	binary.LittleEndian.PutUint32(four[:], uint32(i))
	return bytes.Repeat(four[:], KeySize / 4)
}

func TestTiny(t * testing.T) {
	req := require.New(t)

	m := NewMap(10, 13)
	req.Equal(10, m.maxOccupied)
	req.Equal(12, len(m.buckets))  //15% extra
	req.Equal(m.valueSize, 13)

	//poolSize := m.pool.NumFree()

	//Get when empty
	vbuf := make([]byte, 13)
	req.False(m.Get(int2key(99), vbuf))

	//Fill
	keyInts := []int{67, 38, 41, 75, 77, 27, 50, 3, 19, 91}
	for _, ki := range keyInts {
		res := m.Put(int2key(ki), util.MakeSeq(13, byte(ki + 3)))
		req.Equal(PRKeyWasNew, res)
	}

	//cannot put another
	res := m.Put(int2key(99), make([]byte, 13))
	req.Equal(PRFull, res)

	//Verify all
	for _, ki := range keyInts {
		req.True(m.Get(int2key(ki), vbuf))
		vExpect := util.MakeSeq(13, byte(ki + 3))
		req.Equal(vExpect, vbuf)
	}

	//entire pool was used
	req.Equal(0, m.pool.NumFree())

	//Value which does not exist
	req.False(m.Get(int2key(99), vbuf))

	//TODO: Update every value.
	//Must remove one first because map becomes readonly when completely full.
	//for _, ki := range keyInts {
	//	res := m.Put(int2key(ki), util.MakeSeq(13, byte(ki + 7)))
	//	req.Equal(PRValueUpdated, res)
	//}
}

func TestRand(t * testing.T) {
	req := require.New(t)

	m := NewMap(100, 5)

	N := m.maxOccupied - 1  //one less because map becomes readonly when completely full

	keys := make([][]byte, N)
	vals := make([][]byte, N)
	rand.Seed(1)

	//Fill (except one)
	for i := 0; i < N; i++ {
		keys[i] = randKey()
		vals[i] = randValue(m.valueSize)

		res := m.Put(keys[i], vals[i])
		req.Equal(PRKeyWasNew, res)
	}

	//Verify all
	vbuf := make([]byte, m.valueSize)
	for i := 0; i < N; i++ {
		req.True(m.Get(keys[i], vbuf))
		req.Equal(vals[i], vbuf)
	}

	//entire pool was used (except 1)
	req.Equal(1, m.pool.NumFree())

	//Update all
	for i := 0; i < N; i++ {
		vals[i] = randValue(m.valueSize)
		res := m.Put(keys[i], vals[i])
		req.Equal(PRValueUpdated, res)
	}

	req.Equal(1, m.pool.NumFree())

	//verify
	for i := 0; i < N; i++ {
		req.True(m.Get(keys[i], vbuf))
		req.Equal(vals[i], vbuf)
	}
}

/*
All keys hash to the same bucket and differ only in the last byte
*/
func TestWorstCase(t *testing.T) {
	req := require.New(t)

	m := NewMap(50, 3)

	key := util.MakeSeq(KeySize, 30)
	value := make([]byte, m.valueSize)

	//Fill
	for i := 0; i < 50; i++ {
		key[KeySize - 1] = byte(i)
		util.FillSeq(value, byte(i*3))
		res := m.Put(key, value)
		req.Equal(PRKeyWasNew, res)
	}

	//Verify all
	vbuf := make([]byte, m.valueSize)
	for i := 0; i < 50; i++ {
		key[KeySize - 1] = byte(i)
		util.FillSeq(value, byte(i*3))
		req.True(m.Get(key, vbuf))
		req.Equal(value, vbuf)
	}
}

