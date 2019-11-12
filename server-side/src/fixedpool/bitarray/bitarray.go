/*
A sequential array of bits.  Internally the bits are stored as 64bit integers.
Supports fast searching for zero bits.
*/
package bitarray

type BitArray []uint64

//bits per word
const bpw = 64

//shift by this count to quickly divide or multiply by 64
const bpwShift = 6

//all 64bits ON
const allFF = 0xFFFFFFFFFFFFFFFF

//Represents an invalid bit index
const NotFound = allFF

/*
Allocate a new BitArray.  If nBits will be rounded up to a multiple of 64.
All bits start at 0 (false).
*/
func NewBitArray(nBits uint64) BitArray {
	nWords := nBits >> bpwShift //divide by 64

	//round up
	if (nWords << bpwShift) < nBits || nWords == 0 {
		nWords++
	}

	//check for int overflow (make() requires int)
	nWordsInt := int(nWords)
	if int64(nWordsInt) != int64(nWords) {
		panic("nBits too large")
	}

	return BitArray(make([]uint64, nWords))
}

//Total number of bits
func (ba BitArray) NumBits() uint64 {
	return uint64(len(ba)) * bpw
}

/*
Return which word and which bit within the word
*/
func index2pos(bitIndex uint64) (wordIndex, bitMask uint64) {
	//which word
	wordIndex = bitIndex >> bpwShift  //divide 64
	//which bit within the word
	bitMask = uint64(1) << (bitIndex - (wordIndex << bpwShift))
	return
}

/*Set a bit to 1 (true)*/
func (ba BitArray) Set(bitIndex uint64) {
	wordIndex, bitMask := index2pos(bitIndex)
	ba[wordIndex] |= bitMask
}

/*Set a bit to 0 (false)*/
func (ba BitArray) Clear(bitIndex uint64) {
	wordIndex, bitMask := index2pos(bitIndex)
	ba[wordIndex] &^= bitMask
}

/*Set a bit to 0 but return false if the bit was already zero*/
func (ba BitArray) ClearIfSet(bitIndex uint64) bool {
	wordIndex, bitMask := index2pos(bitIndex)
	if (ba[wordIndex] & bitMask) != 0 {
		ba[wordIndex] &^= bitMask
		return true
	} else {
		//already cleared
		return false
	}
}


func (ba BitArray) IsSet(bitIndex uint64) bool {
	wordIndex, bitMask := index2pos(bitIndex)
	return (ba[wordIndex] & bitMask) != 0
}

//Set all bits to 0
func (ba BitArray) ClearAll() {
	for i := range ba {
		ba[i] = 0
	}
}

//Set all bits to 1 (true)
func (ba BitArray) SetAll() {
	for i := range ba {
		ba[i] = allFF
	}
}

//Set the last N bits to 1.  N cannot exceed 64.
func (ba BitArray) SetLastN(n uint64) {
	if n > bpw {
		panic("SetLastN: n exceeds 64")
	}

	wordIndex := len(ba) - 1
	word := ba[wordIndex]

	bit := uint64(1 << 63)
	for n > 0 {
		word |= bit
		bit >>= 1
		n--
	}

	ba[wordIndex] = word
}


/*
Find the bit index of the first zero bit in the given word.
Returns > bpw if none are zero.
*/
func findZeroBit(word uint64) (bitIndex uint64) {
	/*
	TODO: binary search or something from math/bits (TrailingZeros64?)
	which 32bit half
	which 16bit half
	which 8bit half
	which 4bit half
	which 2bit half
	which bit

	The challenge is keeping track of the bitIndex
	*/

	for bitIndex < bpw && ((word & 1) == 1) {
		word >>= 1
		bitIndex++
	}
	return
}

/*
Linear search for a bit which is 0 (false).  The search is accelerated because it
compares 64bits at a time.

'fromHint' is a bitIndex which can be used to suggest where to start the search.
Searching will wrap arround as need.

Returns NotFound if all bits are 1 (true).
*/
func (ba BitArray) FindZero(fromHint uint64) uint64 {
	startIndex, startMask := index2pos(fromHint)
	endIndex := uint64(len(ba))

	//check the hinted bit (big speedup for sequential allocation)
	if startIndex < endIndex && (ba[startIndex] & startMask) == 0 {
		return fromHint
	}

	var word uint64
	wordIndex := startIndex

	for i := 0; i < 2; i++ {
		for wordIndex < endIndex {
			word = ba[wordIndex]
			if word != allFF {
				//this word has a zero!
				return wordIndex * bpw + findZeroBit(word)
			}
			wordIndex++
		}

		//wrap around
		wordIndex = 0
		endIndex = startIndex
	}

	return NotFound
}
