package gofiberswagger

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"time"
)

/// ------------------------------------------------------------------- ///
/// Functions / constants that are usefull for internal & general usage ///
/// ------------------------------------------------------------------- ///

var (
	timeType       = reflect.TypeOf(time.Time{})
	rawMessageType = reflect.TypeOf(json.RawMessage{})

	zeroInt   = float64(0)
	maxInt    = float64(math.MaxInt)
	minInt    = float64(math.MinInt)
	maxInt8   = float64(math.MaxInt8)
	minInt8   = float64(math.MinInt8)
	maxInt16  = float64(math.MaxInt16)
	minInt16  = float64(math.MinInt16)
	maxInt32  = float64(math.MaxInt32)
	minInt32  = float64(math.MinInt32)
	maxInt64  = float64(math.MaxInt64)
	minInt64  = float64(math.MinInt64)
	maxUint   = float64(math.MaxUint)
	maxUint8  = float64(math.MaxUint8)
	maxUint16 = float64(math.MaxUint16)
	maxUint32 = float64(math.MaxUint32)
	maxUint64 = float64(math.MaxUint64)
)

func replaceNthOccurrence(s, old, new string, n int) string {
	parts := strings.Split(s, old)
	if n <= 0 || n >= len(parts) {
		return s
	}
	result := strings.Join(parts[:n], old) + new + strings.Join(parts[n:], old)
	return result
}
