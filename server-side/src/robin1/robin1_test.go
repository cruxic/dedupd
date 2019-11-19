package robin1

import (
	"testing"
	"github.com/stretchr/testify/require"
	"math/rand"
)

func randKeyValue() (key uint32, value int) {
	key = rand.Uint32()
	value = int(rand.Int31())
	return
}

func TestBasic(t * testing.T) {
	req := require.New(t)

	const capacity = 18
	m := NewMap(capacity)

	req.Equal(20, len(m.buckets))  //~10%
	req.Equal(capacity, m.maxAllowed)

	//Fill to capacity
	rand.Seed(1)
	nPut := 0
	for {
		k, v := randKeyValue()
		if !m.Put(k, v) {
			break
		}
		nPut++
	}

	req.Equal(capacity, nPut)

	//Cannot put another because full
	req.False(m.Put(1234, 5678))

	//t.Log("maxProbeDist", m.maxProbeDist)

	/*fmt.Println("contents")
	for i, b := range m.buckets {
		if b.used {
			fmt.Printf("  %d: %d. prefered %d, dist %d\n", i, b.Key, m.index(b.Key), m.probeDist(b, i))
		} else {
			fmt.Printf("  %d: empty\n", i)
		}
	}*/

	//Verify
	rand.Seed(1)
	for i := 0; i < nPut; i++ {
		k, v1 := randKeyValue()
		v2, found := m.Get(k)
		req.True(found, k)
		req.Equal(v1, v2)
	}

	//Lookup keys which do not exist
	for i := 0; i < 10; i++ {
		k, _ := randKeyValue()
		_, found := m.Get(k)
		req.False(found)
	}
}

type KV struct {
	Key uint32
	Value int
}

func TestExhaustive(t * testing.T) {
	const maxCapacity = 333

	//Generate keys and values
	rand.Seed(1)
	var keys [maxCapacity]KV
	for i := 0; i < maxCapacity; i++ {
		k, v := randKeyValue()
		keys[i].Key = k
		keys[i].Value = v
	}

	//var mpd float32
	//var mpdc, q int

	for capacity := 1; capacity <= maxCapacity; capacity++ {
		m := NewMap(capacity)

		//Fill to capacity
		for i := 0; i < capacity; i++ {
			kv := keys[i]
			if !m.Put(kv.Key, kv.Value) {
				t.Error("put failed")
				return
			}
		}

		//Cannot put another because full
		if m.Put(1234, 5678) {
			t.Error("expected false")
			return
		}

		//Verify
		for i := 0; i < capacity; i++ {
			kv := keys[i]
			v2, found := m.Get(kv.Key)
			if !found {
				t.Error("key not found")
				return
			}
			if kv.Value != v2 {
				t.Error("wrong value")
			}
		}

		//stats := m.CalcStats()
		//t.Log("CP ", stats.CollisionPercent * 100, "AvgProbeDist", stats.AvgProbeDist)

		/*k := m.AvgProbeDist()
		if k > mpd {
			mpd = k
			mpdc = capacity
			q = len(m.buckets)
		}*/

		//t.Log("Probe distance max=", m.maxProbeDist, "avg=", m.AvgProbeDist(), "at", capacity)
	}

	//t.Log(mpd, mpdc, q)
}

//Something about 211 buckets causes high clustering and thus longer average probe distance
func Test211(t * testing.T) {
	req := require.New(t)

	const capacity = 191
	m := NewMap(capacity)

	req.Equal(211, len(m.buckets))
	//m.buckets = make([]Bucket, 210)

	//Fill to capacity
	rand.Seed(1)
	nPut := 0
	for {
		k, v := randKeyValue()
		if !m.Put(k, v) {
			break
		}
		nPut++
		//if nPut > 100 {
		//	stats := m.CalcStats()
		//	fmt.Printf("%g%% full, avg-probe-dist %g\n", stats.PercentUsed * 100.0, stats.AvgProbeDist)
		//}
	}

	//fmt.Printf("avg key %d\n", keysum / uint64(nPut))

	//req.Equal(capacity, nPut)

	/*
	fmt.Println("Stats:")
	stats := m.CalcStats()
	stats.PrintCollisionsPerBucket()
	stats.DebugPrint()
	m.PrintProbeDistances()*/

	/*
	(1.1x8+34)/38 = 12.6% above optimal
	(1.1x8+32)/38 = 7.3% above optimal

	*/
}
