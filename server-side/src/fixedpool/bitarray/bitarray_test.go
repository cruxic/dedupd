package bitarray

import (
	"testing"
	"github.com/stretchr/testify/require"
)


func Test_Basic(t * testing.T) {
	req := require.New(t)

	ba := NewBitArray(190)
	n := ba.NumBits()

	req.Equal(uint64(192), n)  //rounded up

	var i uint64

	//All bits are off
	for i = 0; i < n; i++ {
		req.False(ba.IsSet(i))
	}

	//Set and clear the first bit
	ba.Set(0)
	req.True(ba.IsSet(0))
	ba.Clear(0)
	req.False(ba.IsSet(0))

	//Set and clear the last bit
	last := n - 1
	ba.Set(last)
	req.True(ba.IsSet(last))
	ba.Clear(last)
	req.False(ba.IsSet(last))


	//set even bits
	for i = 0; i < n; i += 2 {
		ba.Set(i)
	}

	//verify
	for i = 0; i < n; i += 2 {
		req.True(ba.IsSet(i))
		req.False(ba.IsSet(i+1))
	}

	ba.ClearAll()

	//verify
	for i = 0; i < n; i++ {
		req.False(ba.IsSet(i))
	}

	//set odd bits
	for i = 1; i < n; i += 2 {
		ba.Set(i)
	}

	//verify
	for i = 1; i < n; i += 2 {
		req.False(ba.IsSet(i-1))
		req.True(ba.IsSet(i))
	}
}

func Test_Find(t * testing.T) {
	req := require.New(t)

	ba := NewBitArray(192)  //3 words
	n := ba.NumBits()

	req.Equal(uint64(0), ba.FindZero(0))
	req.Equal(uint64(63), ba.FindZero(63))
	req.Equal(uint64(64), ba.FindZero(64))
	req.Equal(uint64(191), ba.FindZero(191))
	req.Equal(uint64(0), ba.FindZero(999))  //wrapped

	//Set all but 1 and then find it
	ba.SetAll()
	ba.Clear(99)
	req.Equal(uint64(99), ba.FindZero(0))
	req.Equal(uint64(99), ba.FindZero(180))
	ba.Set(99)

	//Try finding every bit
	var i uint64
	for i = 0; i < n; i++ {
		ba.Clear(i)
		//find from zero
		req.Equal(i, ba.FindZero(0))

		//find from 3 before
		if i > 3 {
			req.Equal(i, ba.FindZero(i-3))
		}

		//find from 3 after
		req.Equal(i, ba.FindZero(i+3))

		ba.Set(i)
	}
}

func Test_ClearIfSet(t * testing.T) {
	req := require.New(t)

	ba := NewBitArray(192)  //3 words

	ba.Set(99)
	req.True(ba.ClearIfSet(99))
	req.False(ba.ClearIfSet(99))
}

func Test_SetLastN(t * testing.T) {
	req := require.New(t)

	ba := NewBitArray(192)  //3 words
	n := ba.NumBits()

	//last 0
	ba.SetLastN(0)
	req.False(ba.IsSet(n - 1))

	//last 1
	ba.SetLastN(1)
	req.True(ba.IsSet(n - 1))
	ba.Clear(n - 1)

	//last 13
	ba.SetLastN(13)

	//verify last 13
	var i uint64
	for i = 0; i < n - 13; i++ {
		req.False(ba.IsSet(i))
	}
	for i < n {
		req.True(ba.IsSet(i))
		i++
	}

	//last 64
	ba.ClearAll()

	ba.SetLastN(64)

	for i = 0; i < n - 64; i++ {
		req.False(ba.IsSet(i))
	}
	for i < n {
		req.True(ba.IsSet(i))
		i++
	}

}
