/*
Playground to experiment with robin hood hashing.

https://www.sebastiansylvan.com/post/robin-hood-hashing-should-be-your-default-hash-table-implementation/

http://codecapsule.com/2013/11/17/robin-hood-hashing-backward-shift-deletion/
*/
package robin1

import (
	"fmt"
)


type Bucket struct {
	Key uint32
	Value int
	used bool
}

type Map struct {
	buckets []Bucket
	nUsed int
	//90% of len(buckets)
	maxAllowed int
}

func NewMap(maxCapacity int) *Map {
	if maxCapacity < 1 {
		maxCapacity = 1
	}

	//10% extra
	nBuckets := int(float32(maxCapacity) * 1.10) + 1

	return &Map{
		buckets: make([]Bucket, nBuckets),
		maxAllowed: maxCapacity,
	}
}

func (m *Map) index(key uint32) int {
	//for now we assume key == hashcode
	return int(key % uint32(len(m.buckets)))
}

func (m *Map) probeDist(b Bucket, currentIndex int) int {
	desiredIndex := m.index(b.Key)
	if currentIndex < desiredIndex {
		//wrapped due to modulo
		return len(m.buckets) - desiredIndex + currentIndex
	} else {
		return currentIndex - desiredIndex
	}
}

type Stats struct {
	NumBuckets int
	NumUsed int
	NumEmpty int
	MaxAllowed int
	AvgProbeDist float32
	MaxProbeDist int
	//Percentage of buckets which are occupied
	PercentUsed float32
	//Percentage of NumUsed which hash to the same bucket (0.0 to 1.0)
	CollisionPercent float32
	//For each bucket, the count of keys which hash to that bucket.
	CollisionsPerBucket []int
}

func (stats Stats) DebugPrint() {
	fmt.Printf("  NumBuckets: %d\n", stats.NumBuckets)
	fmt.Printf("  NumUsed: %d\n", stats.NumUsed)
	fmt.Printf("  NumEmpty: %d\n", stats.NumEmpty)
	fmt.Printf("  PercentUsed: %g\n", stats.PercentUsed * 100.0)
	fmt.Printf("  MaxAllowed: %d\n", stats.MaxAllowed)
	fmt.Printf("  MaxProbeDist: %d\n", stats.MaxProbeDist)
	fmt.Printf("  AvgProbeDist: %g\n", stats.AvgProbeDist)
	fmt.Printf("  CollisionPercent: %g\n", stats.CollisionPercent * 100.0)
}

func (stats Stats) PrintCollisionsPerBucket() {
	dots := "********"
	fmt.Printf("Bucket: Num-Occupants\n")
	for bucketIdx, count := range stats.CollisionsPerBucket {
		for len(dots) < count {
			dots = dots + dots
		}
		fmt.Printf("% 6d: %s\n", bucketIdx, dots[0:count])
	}
}

func (m *Map) CalcStats() Stats {
	var sum int64
	var stats Stats

	collisionCounts := make(map[int]int)

	for idx, bucket := range m.buckets {
		if bucket.used {
			stats.NumUsed++
			dist := m.probeDist(bucket, idx)
			sum += int64(dist)
			if dist > stats.MaxProbeDist {
				stats.MaxProbeDist = dist
			}

			collisionCounts[m.index(bucket.Key)]++
		} else {
			stats.NumEmpty++
		}
	}

	if stats.NumUsed != m.nUsed {
		panic("nUsed mismatch")
	}

	stats.AvgProbeDist = float32(float64(sum) / float64(stats.NumUsed))

	stats.MaxAllowed = m.maxAllowed
	stats.NumBuckets = len(m.buckets)
	stats.PercentUsed = float32(stats.NumUsed) / float32(stats.NumBuckets)

	totalCollisions := 0
	for _, count := range collisionCounts {
		if count > 1 {
			totalCollisions += count
		}
	}
	stats.CollisionPercent = float32(totalCollisions) / float32(stats.NumUsed)

	stats.CollisionsPerBucket = make([]int, stats.NumBuckets)
	for i := range stats.CollisionsPerBucket {
		stats.CollisionsPerBucket[i] = collisionCounts[i]
	}

	return stats
}

func (m *Map) PrintProbeDistances() {
	fmt.Println("Bucket: Occupant distance from optimal")
	for idx, bucket := range m.buckets {
		if bucket.used {
			dist := m.probeDist(bucket, idx)
			fmt.Printf("% 6d: %d\n", idx, dist)
		} else {
			fmt.Printf("% 6d:\n", idx)
		}
	}
}

func (m *Map) Put(key uint32, value int) bool {
	if m.nUsed >= m.maxAllowed {
		//we are 90% full.
		//TODO: this disallows updating of existing entries when full
		return false
	}

	incoming := Bucket{
		Key: key,
		Value: value,
		used: true,
	}

	idx := m.index(key)
	dist := 0
	wrapCount := 0

	for {
		other := m.buckets[idx]

		if !other.used {
			//Use empty bucket
			m.nUsed++
			m.buckets[idx] = incoming
			return true
		} else if other.Key == incoming.Key {
			//Update existing
			m.buckets[idx].Value = incoming.Value
			return true
		}

		otherDist := m.probeDist(other, idx)
		if otherDist < dist {
			//Rob from the rich and give to the poor!
			m.buckets[idx] = incoming
			incoming = other
			dist = otherDist
		}

		dist++
		//if dist > m.maxProbeDist {
		//	m.maxProbeDist = dist
		//}

		idx++
		if idx >= len(m.buckets) {
			idx = 0

			wrapCount++
			if wrapCount > 1 {
				//will not happen because we disallow Put when nUsed >= maxAllowed
				panic("Put assertion fail")
			}
		}
	}
}

func (m *Map) Get(key uint32) (value int, found bool) {
	idx := m.index(key)
	dist := 0
	wrapped := false

	for {
		other := m.buckets[idx]

		if !other.used {
			//not found
			break
		} else if other.Key == key {
			//Found!
			found = true
			value = other.Value
			break
		}

		otherDist := m.probeDist(other, idx)
		if otherDist < dist {
			//not found
			break
		}

		dist++
		idx++
		if idx >= len(m.buckets) {
			if wrapped {
				//not found
				break
			}
			idx = 0
			wrapped = true
		}
	}

	return
}
