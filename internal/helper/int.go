package helper

import "reflect"

// Add1 returns the value of a + 1.
// The type of a must be an integer type or the result is 0.
func Add1(a any) int64 {
	switch val := a.(type) {
	case
		int,
		int8,
		int16,
		int32,
		int64:
		i := reflect.ValueOf(val).Int()
		return i + 1
	default:
		return 0
	}
}
