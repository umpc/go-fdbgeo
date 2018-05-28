// Package fdbgeo contains tools for building geospatial layers using FoundationDB
// with Geohash-encoded keys.
//
// This package uses its complimentary "zrange" package: https://github.com/umpc/go-zrange,
// for performing geospatial range queries using FoundationDB with Geohash-encoded
// keys and a search radius.
//
// The `RadialRange` method appears to be sufficient for range queries of around
// 5,000km or less. Changes that efficiently add support for larger query ranges
// are welcome here: https://github.com/umpc/go-zrange.
//
package fdbgeo
