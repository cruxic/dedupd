package util


import (
	"testing"
	"github.com/stretchr/testify/require"
)


func Test_utils(t * testing.T) {
	req := require.New(t)

	bb := make([]byte, 8)
	Uint64ToBytes(0x7d43670ae96abcd1, bb)
	req.Equal(bb[0], byte(0xd1))

	val := Uint64FromBytes(bb)
	req.Equal(val, uint64(0x7d43670ae96abcd1))

	Uint24ToBytes(0xd5c8a9, bb)
	req.Equal(bb[0], byte(0xa9))

	val24 := Uint24FromBytes(bb)
	req.Equal(uint32(0xd5c8a9), val24)

	Uint32ToBytes(0xe96abcd1, bb)
	req.Equal(uint32(0xe96abcd1), Uint32FromBytes(bb))

}

func Test_rand(t * testing.T) {
	req := require.New(t)

	rs := NewRandStream(0x11)

	m := make(map[uint32]bool)

	for i := 0; i < 1000; i++ {
		r := rs.Rand24bit()
		req.False(m[r], i)  //not yet present
		m[r] = true
	}

	//different seeds produce different values
	rs1 := NewRandStream(0x11)
	rs2 := NewRandStream(0x12)

	req.True(rs1.Rand24bit() != rs2.Rand24bit())
}

