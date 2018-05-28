package fdbgeo

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/mmcloughlin/geohash"
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
	if params.BitsOfPrecision == 0 {
		params.BitsOfPrecision = 64
	}
	if params.Subspace == nil {
		params.Subspace = subspace.FromBytes(nil)
	}
	return params
}

// WithinRadius determines if a Geohash is within the specified radius.
// Running WithinRadius in a RangeIterator loop may double transaction time though
// make parsing more efficient. Its potential benefits are dependent on the data
// model in use.
func (params RadialRangeParams) WithinRadius(geohashID uint64) bool {
	params = params.setDefaults()

	latitude, longitude := geohash.DecodeIntWithPrecision(
		geohashID,
		params.BitsOfPrecision,
	)
	distanceKm := zrange.Haversine(
		params.Latitude, params.Longitude,
		latitude, longitude,
	)

	return distanceKm < params.Radius
}
