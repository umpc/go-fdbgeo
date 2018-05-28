package fdbgeo

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/umpc/go-zrange"
)

func createKeyRanges(sub subspace.Subspace, hashRangeList zrange.HashRanges) []fdb.KeyRange {
	keyRangeList := make([]fdb.KeyRange, len(hashRangeList))

	for i, hashRange := range hashRangeList {
		keyRangeList[i] = fdb.KeyRange{
			Begin: sub.Sub(hashRange.Min),
			End:   sub.Sub(hashRange.Max),
		}
	}

	return keyRangeList
}
