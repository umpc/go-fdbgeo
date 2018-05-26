package fdbgeo

import (
	"math"
	"sort"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/mmcloughlin/geohash"
)

// RadialRange uses a radius in kilometers, a latitude, and a longitude to return
// a slice of one or more ranges of keys that can be used to efficiently perform
// geohash-based spatial queries.
//
// This method uses an algorithm that was derived from the "Search" section of this page:
// https://web.archive.org/web/20180526044934/https://github.com/yinqiwen/ardb/wiki/Spatial-Index#search
//
// RadialRange expands upon the ideas referenced above, by:
//
// • Sorting key ranges
//
// • Combining overlapping key ranges
//
// • Handling overflows resulting from bitshifting, such as when querying for: (-90, -180)
//
func RadialRange(params RadialRangeParams) []fdb.KeyRange {
	return params.
		setDefaults().
		findNeighboringRanges().
		combineRanges().
		createKeyRanges(params.Subspace)
}

// RadialRangeParams defaults to expecting 64-bit geohash-encoded keys.
// A subspace will be prepended if one is set.
type RadialRangeParams struct {
	BitsOfPrecision uint
	Radius,
	Latitude,
	Longitude float64
	Subspace subspace.Subspace
}

func (params RadialRangeParams) setDefaults() RadialRangeParams {
	if params.Subspace == nil {
		params.Subspace = subspace.FromBytes(nil)
	}
	if params.BitsOfPrecision == 0 {
		params.BitsOfPrecision = 64
	}
	return params
}

func (params RadialRangeParams) radiusToBits() uint {
	const initialSignificantBits = 2

	for i := len(radiusToBits) - 1; i > 0; i-- {
		if params.Radius < radiusToBits[i] {
			return uint(i*2 + initialSignificantBits)
		}
	}

	return uint(initialSignificantBits)
}

func (params RadialRangeParams) findNeighboringRanges() hashRanges {
	rangeBits := params.radiusToBits()

	queryPoint := geohash.EncodeIntWithPrecision(
		params.Latitude,
		params.Longitude,
		rangeBits,
	)

	neighborList := neighbors(geohash.NeighborsIntWithPrecision(queryPoint, rangeBits))
	neighborList = append(neighborList, queryPoint)

	diff := params.BitsOfPrecision - rangeBits
	return neighborList.expandRanges(diff)
}

const (
	earthSemiMajorAxis = 6378.137
	earthEquator       = math.Pi * earthSemiMajorAxis
)

var radiusToBits = precalcRadiusToBits()

func precalcRadiusToBits() []float64 {
	var radiusToBits []float64

	for bits, prevRadialBound := uint(4), earthEquator; bits < 64; bits += 2 {
		radiusToBits = append(radiusToBits, prevRadialBound/2)
		prevRadialBound = radiusToBits[len(radiusToBits)-1]
	}

	return radiusToBits
}

type hashRange struct {
	Min, Max uint64
}

type hashRangesMinAscSorter []hashRange

func (s hashRangesMinAscSorter) Len() int {
	return len(s)
}
func (s hashRangesMinAscSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s hashRangesMinAscSorter) Less(i, j int) bool {
	return s[i].Min < s[j].Min
}

type neighbors []uint64

func (neighborList neighbors) expandRanges(rangeBitsDiff uint) hashRanges {
	hashRangeList := make(hashRanges, 0, len(neighborList))

	for _, neighbor := range neighborList {
		min := neighbor << rangeBitsDiff
		max := (neighbor + 1) << rangeBitsDiff

		// Handle overflows near the outer edges.
		// Example: (-90.0, -180.0)
		if min > max {
			continue
		}

		hashRangeList = append(hashRangeList, hashRange{
			Min: min,
			Max: max,
		})
	}

	return hashRangeList
}

type hashRanges []hashRange

func (hashRangeList hashRanges) combineRanges() hashRanges {
	sort.Sort(hashRangesMinAscSorter(hashRangeList))
	combinedHashRangeList := hashRangeList[:0]

	for i := 0; i < len(hashRangeList)-1; i++ {
		hashRange := hashRangeList[i]
		nextHashRange := hashRangeList[i+1]

		if hashRange.Max == nextHashRange.Min {
			hashRange.Max = nextHashRange.Max
		}

		if hashRange.Max == nextHashRange.Max {
			hashRangeList[i+1].Min = hashRange.Min
			continue
		}

		combinedHashRangeList = append(combinedHashRangeList, hashRange)
	}

	return append(combinedHashRangeList, hashRangeList[len(hashRangeList)-1])
}

func (hashRangeList hashRanges) createKeyRanges(sub subspace.Subspace) []fdb.KeyRange {
	keyRangeList := make([]fdb.KeyRange, 0, len(hashRangeList))

	for _, hashRange := range hashRangeList {
		keyRangeList = append(keyRangeList, fdb.KeyRange{
			Begin: sub.Sub(hashRange.Min),
			End:   sub.Sub(hashRange.Max),
		})
	}

	return keyRangeList
}
