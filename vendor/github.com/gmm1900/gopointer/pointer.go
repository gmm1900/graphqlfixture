package gopointer

import "time"

// OfString returns the pointer to the given string
func OfString(str string) *string {
	return &str
}

// OfInt returns the pointer to the given int
func OfInt(i int) *int {
	return &i
}

// OfInt64 returns the pointer to the given int64
func OfInt64(i64 int64) *int64 {
	return &i64
}

// OfFloat32 returns the pointer to the given float32
func OfFloat32(f float32) *float32 {
	return &f
}

// OfFloat64 returns the pointer to the given float64
func OfFloat64(f float64) *float64 {
	return &f
}

// OfBool returns the pointer to the given bool
func OfBool(b bool) *bool {
	return &b
}

// OfTime returns the pointer to the given time
func OfTime(t time.Time) *time.Time {
	return &t
}

// OfNilTime returns nil value of the *time.Time type
func OfNilTime() *time.Time {
	return nil
}

// OfNilJSON returns the pointer to a nil JSON bytes
func OfNilJSON() *[]byte {
	return nil
}