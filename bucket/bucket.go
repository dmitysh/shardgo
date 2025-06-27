package bucket

import "strconv"

// Bucket bucket
type Bucket uint

// Range range of Bucket
type Range struct {
	From Bucket
	To   Bucket
}

// String bucket name
func (s Bucket) String() string {
	return strconv.Itoa(int(s)) //nolint: gosec
}
