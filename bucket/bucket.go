package bucket

import "strconv"

type Bucket uint

type Range struct {
	From Bucket
	To   Bucket
}

func (s Bucket) String() string {
	return strconv.Itoa(int(s))
}
