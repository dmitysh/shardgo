package shard

import (
	"github.com/dmitysh/shardgo/bucket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spaolacci/murmur3"
)

type DSN string

type Shard struct {
	DSN
	Pool *pgxpool.Pool
}

type BucketToShard struct {
	ShardDSN    DSN
	BucketsList []bucket.Range
}

type KeyToBucketFunc func(key string) bucket.Bucket

func HashingKeyToBucket(bucketsCount int) KeyToBucketFunc {
	return func(key string) bucket.Bucket {
		hash := murmur3.Sum64([]byte(key))
		bucketNumber := hash % uint64(bucketsCount)
		return bucket.Bucket(bucketNumber)
	}
}
