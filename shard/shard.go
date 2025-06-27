package shard

import (
	"github.com/dmitysh/shardgo/bucket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spaolacci/murmur3"
)

// DSN shard's dsn
type DSN string

// Shard shard
type Shard struct {
	DSN
	Pool *pgxpool.Pool
}

// BucketToShard mapping for range of buckets to shard
type BucketToShard struct {
	ShardDSN    DSN
	BucketsList []bucket.Range
}

// KeyToBucketFunc determines bucket by key
type KeyToBucketFunc func(key string) bucket.Bucket

// HashingKeyToBucket standard hashing KeyToBucketFunc func
func HashingKeyToBucket(bucketsCount int) KeyToBucketFunc {
	return func(key string) bucket.Bucket {
		hash := murmur3.Sum64([]byte(key))
		bucketNumber := hash%uint64(bucketsCount) + 1 //nolint: gosec
		return bucket.Bucket(bucketNumber)
	}
}
