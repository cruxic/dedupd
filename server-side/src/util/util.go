package util

import (
	"log"
	"crypto/cipher"
	"crypto/aes"
)

func FillConst(dest []byte, byteVal byte) {
	for i := range dest {
		dest[i] = byteVal
	}
}

func FillSeq(dest []byte, startValue byte) {
	for i := range dest {
		dest[i] = byte(int(startValue) + i)
	}
}

func MakeSeq(count int, startValue byte) []byte {
	dat := make([]byte, count)
	FillSeq(dat, startValue)
	return dat
}


type RandStream struct {
	stream cipher.Stream
}

func NewRandStream(seed byte) RandStream {
	iv := make([]byte, aes.BlockSize)
	key := make([]byte, aes.BlockSize)
	FillConst(iv, seed)
	FillConst(key, 0xA7)

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	return RandStream{
		stream: cipher.NewCTR(blockCipher, iv),
	}
}

func (rs RandStream) FillRand(dest []byte) {
	rs.stream.XORKeyStream(dest, dest)
}

func (rs RandStream) Rand24bit() uint32 {
	var temp [3]byte
	tempSlice := temp[:]
	rs.stream.XORKeyStream(tempSlice, tempSlice)
	return Uint24FromBytes(tempSlice)
}

func (rs RandStream) Rand16bit() uint16 {
	var temp [2]byte
	tempSlice := temp[:]
	rs.stream.XORKeyStream(tempSlice, tempSlice)
	return uint16(temp[0]) | (uint16(temp[1]) << 8)
}


func (rs RandStream) Rand32bit() uint32 {
	var temp [4]byte
	tempSlice := temp[:]
	rs.stream.XORKeyStream(tempSlice, tempSlice)
	return Uint32FromBytes(tempSlice)
}

//Read a 24bit little-endian value
func Uint24FromBytes(src []byte) uint32 {
	return uint32(src[0]) |
		(uint32(src[1]) << 8) |
		(uint32(src[2]) << 16)
}

func Uint24ToBytes(val uint32, dest []byte) {
	dest[0] = byte(val & 0xFF)
	dest[1] = byte((val >> 8) & 0xFF)
	dest[2] = byte((val >> 16) & 0xFF)
	dest[3] = byte((val >> 24) & 0xFF)
}

//Read a 32bit little-endian value
func Uint32FromBytes(src []byte) uint32 {
	return uint32(src[0]) |
		(uint32(src[1]) << 8) |
		(uint32(src[2]) << 16) |
		(uint32(src[3]) << 24)
}

func Uint32ToBytes(val uint32, dest []byte) {
	dest[0] = byte(val & 0xFF)
	dest[1] = byte((val >> 8) & 0xFF)
	dest[2] = byte((val >> 16) & 0xFF)
	dest[3] = byte((val >> 24) & 0xFF)
}

//Read a 64bit little-endian value
func Uint64FromBytes(src []byte) uint64 {
	return uint64(src[0]) |
		(uint64(src[1]) << 8) |
		(uint64(src[2]) << 16) |
		(uint64(src[3]) << 24) |
		(uint64(src[4]) << 32) |
		(uint64(src[5]) << 40) |
		(uint64(src[6]) << 48) |
		(uint64(src[7]) << 56)
}

func Uint64ToBytes(val uint64, dest []byte) {
	dest[0] = byte(val & 0xFF)
	dest[1] = byte((val >> 8) & 0xFF)
	dest[2] = byte((val >> 16) & 0xFF)
	dest[3] = byte((val >> 24) & 0xFF)
	dest[4] = byte((val >> 32) & 0xFF)
	dest[5] = byte((val >> 40) & 0xFF)
	dest[6] = byte((val >> 48) & 0xFF)
	dest[7] = byte((val >> 56) & 0xFF)
}
