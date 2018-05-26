package fdbgeo

import (
	"github.com/umpc/go-zrange"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
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
	params = params.setDefaults()

	hashRangeList := zrange.RadialRange(zrange.RadialRangeParams{
		BitsOfPrecision: params.BitsOfPrecision,
		Radius:          params.Radius,
		Latitude:        params.Latitude,
		Longitude:       params.Longitude,
	})

	return createKeyRanges(params.Subspace, hashRangeList)
}

// RadialRangeParams specifies arguments for the RadialRange method.
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
	return params
}

func createKeyRanges(sub subspace.Subspace, hashRangeList zrange.HashRanges) []fdb.KeyRange {
	keyRangeList := make([]fdb.KeyRange, 0, len(hashRangeList))

	for _, hashRange := range hashRangeList {
		keyRangeList = append(keyRangeList, fdb.KeyRange{
			Begin: sub.Sub(hashRange.Min),
			End:   sub.Sub(hashRange.Max),
		})
	}

	return keyRangeList
}
