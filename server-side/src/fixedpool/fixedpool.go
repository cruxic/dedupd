/*
An allocator of fixed size blocks of bytes.  An allocated
block has 32bit pointer.  This saves memory compared to 64bit pointers.
Internal overhead is only 1bit per block.

Random alloc/free traffic is probably slower than Go's heap but
with the advantage of nearly zero overhead per allocation and no garbage
collection pauses.
*/
package fixedpool

import (
	"fixedpool/bitarray"
)

//32bit pointer.  A valid allocation is never zero.
type Ptr uint32

const Zero Ptr = 0

type Pool struct {
	blockSize int
	nUsed int
	data []byte
	//1 bit per block to track which have been allocated
	allocMask bitarray.BitArray
	nextAllocIndex uint64
}

func NewPool(blockSize, numBlocks int) *Pool {
	if numBlocks <= 0 || blockSize <= 0 {
		panic("NewPool illegal arg")
	}

	pool := &Pool {
		blockSize: blockSize,
		data: make([]byte, numBlocks * blockSize),
		allocMask: bitarray.NewBitArray(uint64(numBlocks)),
	}

	//BitArray rounds up to a multiple of 64.  Mark these all allocated.
	pool.allocMask.SetLastN(pool.allocMask.NumBits() - uint64(numBlocks))

	return pool
}

func (pool *Pool) NumBlocks() int {
	return len(pool.data) / pool.blockSize
}

func (pool *Pool) BlockSize() int {
	return pool.blockSize
}

func (pool *Pool) NumUsed() int {
	return pool.nUsed
}

func (pool *Pool) NumFree() int {
	return pool.NumBlocks() - pool.nUsed
}


/*
Access the block at the given Ptr.  This function thread-safe as long
as each Ptr is only accessed by one thread.

This function DOES NOT check if the given Ptr has already been freed.
*/
func (pool *Pool) Get(ptr Ptr) []byte {
	if ptr == Zero {
		panic("fixedpool.Fetch: zero ptr")
	}
	bs := uint32(pool.blockSize)
	offset := (uint32(ptr) - 1) * bs
	return pool.data[offset:offset+bs]
}

/*
Allocate one block.  Returns 0 if no free blocks.
*/
func (pool *Pool) Alloc() Ptr {
	freeIndex := pool.allocMask.FindZero(pool.nextAllocIndex)
	if freeIndex == bitarray.NotFound {
		return Zero
	} else {
		pool.allocMask.Set(freeIndex)
		pool.nextAllocIndex = freeIndex + 1
		pool.nUsed++
		return Ptr(uint32(freeIndex) + 1)  //+1 ensures never zero
	}
}

//fill given slice with zeros
func fillZero(dest []byte) {
	var _zeros [128]byte
	zeros := _zeros[:]

	i := 0
	for i < len(dest) {
		copy(dest[i:], zeros)
		i += len(_zeros)
	}
}

/*
Return a block to the pool.  Harmless to pass zero.
Passing an already freed Ptr will panic.
*/
func (pool *Pool) Free(ptr Ptr) {
	if ptr != Zero {
		//Clear it to zero so that it's ready for reallocation.
		fillZero(pool.Get(ptr))

		index := uint64(ptr) - 1
		if !pool.allocMask.ClearIfSet(index) {
			panic("fixedpool.Free: already freed")
		}

		pool.nUsed--
		pool.nextAllocIndex = index
	}
}

/*
Reset the pool.
*/
func (pool *Pool) FreeAll() {
	pool.allocMask.ClearAll()
	fillZero(pool.data)
	pool.nUsed = 0
	pool.nextAllocIndex = 0
}
