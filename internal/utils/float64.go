// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Float64ToBytes converts a float64 to a 8-byte big-endian byte slice.
func Float64ToBytes(f float64) []byte {
	bits := math.Float64bits(f)
	// For lexicographical sorting of positive floats in bbolt:
	// We use BigEndian to ensure that larger bit representations
	// (which generally correlate to larger positive floats) come after.
	// Note: This simple approach only works for non-negative floats.
	// For negative floats, we'd need more complex bit manipulation.
	// Since cue numbers are expected to be positive, this should suffice.
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, bits)
	return buf
}

// BytesToFloat64 converts a 8-byte big-endian byte slice to a float64.
func BytesToFloat64(b []byte) (float64, error) {
	if len(b) != 8 {
		return 0, fmt.Errorf("invalid byte length for float64: %d", len(b))
	}
	bits := binary.BigEndian.Uint64(b)
	return math.Float64frombits(bits), nil
}
