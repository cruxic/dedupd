package fixedpool

import (
	"testing"
	"github.com/stretchr/testify/require"
	"util"
)

func isAllZero(dat []byte) bool {
	for _, b := range dat {
		if b != 0 {
			return false
		}
	}
	return true
}

func Test_fillZero(t * testing.T) {
	dat := make([]byte, 99999)
	for i := range dat {
		dat[i] = byte(0x99)
	}

	fillZero(dat)

	for i := range dat {
		if dat[i] != 0x00 {
			t.Error("not zero!")
		}
	}
}

func Test_Basic(t * testing.T) {
	req := require.New(t)

	const blockSize = 7
	const nBlocks = 13
	pool := NewPool(blockSize, nBlocks)

	req.Equal(nBlocks, pool.NumFree())
	req.Equal(0, pool.NumUsed())

	//Alloc all and write to each block
	var pointers [nBlocks]Ptr
	for i := 0; i < nBlocks; i++ {
		pointers[i] = pool.Alloc()
		req.True(pointers[i] != Zero)
		req.Equal(i + 1, pool.NumUsed())
	}

	//no more
	req.True(pool.Alloc() == Zero)
	req.True(pool.Alloc() == Zero)

	//write to each block
	for i, ptr := range pointers {
		dat := pool.Get(ptr)
		req.Equal(blockSize, len(dat))
		util.FillConst(dat, byte(i + 1))
	}

	//verify each block
	for i, ptr := range pointers {
		dat := pool.Get(ptr)
		req.Equal(blockSize, len(dat))
		b := byte(i + 1)
		req.Equal(b, dat[0])
		req.Equal(b, dat[blockSize - 1])
	}

	//Free the odd blocks
	for i := 1; i < nBlocks; i += 2 {
		pool.Free(pointers[i])
		pointers[i] = Zero
	}

	//Allocate them again.  Verify each holds zeros
	n := nBlocks / 2
	j := 1
	for i := 0; i < n; i++ {
		ptr := pool.Alloc()
		req.True(ptr != Zero)

		req.True(pointers[j] == Zero)
		pointers[j] = ptr

		dat := pool.Get(ptr)
		req.True(isAllZero(dat))
		util.FillConst(dat, byte(j + 1))
		j += 2
	}
	req.True(pool.Alloc() == Zero)

	//verify each block (again)
	for i, ptr := range pointers {
		dat := pool.Get(ptr)
		req.Equal(blockSize, len(dat))
		b := byte(i + 1)
		req.Equal(b, dat[0])
		req.Equal(b, dat[blockSize - 1])
	}

	req.Equal(0, pool.NumFree())
	req.Equal(nBlocks, pool.NumUsed())
}

func Test_FreeAll(t *testing.T) {
	req := require.New(t)

	const blockSize = 7
	const nBlocks = 13
	pool := NewPool(blockSize, nBlocks)

	blocks := make([][]byte, nBlocks)

	for j := 0; j < 3; j++ {
		for i := 0; i < nBlocks; i++ {
			ptr := pool.Alloc()
			req.True(ptr != Zero)
			block := pool.Get(ptr)
			req.Equal(blockSize, len(block))
			util.FillConst(block, byte(i + 1))
			blocks[i] = block
		}

		pool.FreeAll()

		//verify all blocks were reset to zero
		for _, block := range blocks {
			req.True(isAllZero(block))
		}
	}

}

func benchSequentialAlloc(pool *Pool) bool {
	nBlocks := pool.NumBlocks()
	pool.FreeAll()

	for i := 0; i < nBlocks; i++ {
		ptr := pool.Alloc()
		if ptr == Zero {
			return false
		}
	}

	return true
}

func Benchmark_sequentialAlloc(b *testing.B) {
	const blockSize = 3
	const nBlocks = 1000000
	pool := NewPool(blockSize, nBlocks)

	for j := 0; j < b.N; j++ {
		if !benchSequentialAlloc(pool) {
			panic("alloc fail")
		}
	}
}
