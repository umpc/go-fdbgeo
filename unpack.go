package fdbgeo

import (
	"fmt"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

// UnpackUint parses and returns a uint64 value from a key. A negative idx value
// is treated as relative to the end of the parsed key slice.
func UnpackUint(key fdb.Key, idx int) (uint64, error) {
	k, err := tuple.Unpack(key)
	if err != nil {
		return 0, nil
	}

	if idx < 0 {
		idx += len(k)
	}

	const errFmt = "idx is %s index of the parsed key"
	if idx < 0 {
		return 0, fmt.Errorf(errFmt, "less than the first")
	}
	if idx > len(k)-1 {
		return 0, fmt.Errorf(errFmt, "greater than the final")
	}

	var ret uint64
	switch v := k[idx].(type) {
	case int64:
		ret = uint64(v)
	case uint64:
		ret = v
	}

	return ret, nil
}
