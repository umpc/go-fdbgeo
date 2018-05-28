package fdbgeo

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/umpc/go-zrange"
)

// RadialRange uses a radius in kilometers, a latitude, and a longitude to return
// a slice of one or more ranges of keys that can be used to efficiently perform
// Geohash-based spatial queries.
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
func (params RadialRangeParams) RadialRange() []fdb.KeyRange {
	params = params.setDefaults()

	hashRangeList := params.zrange().RadialRange()
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

// WithinRadius determines whether a Geohash is within the specified radius.
// Running WithinRadius in a RangeIterator loop may double transaction time though
// make parsing more efficient. Its potential benefits are dependent on the data
// model in use.
func (params RadialRangeParams) WithinRadius(geohashID uint64) bool {
	return params.zrange().WithinRadius(geohashID)
}

func (params RadialRangeParams) setDefaults() RadialRangeParams {
	if params.Subspace == nil {
		params.Subspace = subspace.FromBytes(nil)
	}
	return params
}

func (params RadialRangeParams) zrange() zrange.RadialRangeParams {
	return zrange.RadialRangeParams{
		BitsOfPrecision: params.BitsOfPrecision,
		Radius:          params.Radius,
		Latitude:        params.Latitude,
		Longitude:       params.Longitude,
	}
}
