package shardgo

import (
	"context"
	"fmt"
	"github.com/dmitysh/shardgo/bucket"
	"golang.org/x/sync/errgroup"
	"slices"

	"github.com/dmitysh/shardgo/shard"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ShardCluster struct {
	keyToBucket   shard.KeyToBucketFunc
	bucketToShard map[bucket.Bucket]shard.Shard
	shards        []shard.Shard
}

func NewShardCluster(ctx context.Context, keyToBucket shard.KeyToBucketFunc, shardList []shard.BucketToShard) (*ShardCluster, error) {
	bucketToShard := make(map[bucket.Bucket]shard.Shard)

	shards := make([]shard.Shard, 0)
	for _, s := range shardList {
		pool, err := pgxpool.New(ctx, string(s.ShardDSN))
		if err != nil {
			return nil, fmt.Errorf("can't create pool: %w", err)
		}

		for _, buck := range getBucketsFromList(s.BucketsList) {
			newShard := shard.Shard{
				DSN:  s.ShardDSN,
				Pool: pool,
			}
			bucketToShard[buck] = newShard
			if !slices.ContainsFunc(shards, func(s shard.Shard) bool {
				return s.DSN == newShard.DSN
			}) {
				shards = append(shards, newShard)
			}
		}
	}

	sc := &ShardCluster{
		keyToBucket:   keyToBucket,
		bucketToShard: bucketToShard,
		shards:        shards,
	}

	err := sc.ForEachShard(ctx, func(ctx context.Context, sh shard.Shard) error {
		return sh.Pool.Ping(ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("can't ping shard: %w", err)
	}

	return sc, nil
}

func getBucketsFromList(buckets []bucket.Range) []bucket.Bucket {
	flattenedBuckets := make([]bucket.Bucket, 0)
	for _, b := range buckets {
		for bucketID := b.From; bucketID < b.To; bucketID++ {
			flattenedBuckets = append(flattenedBuckets, bucketID)
		}
	}
	return flattenedBuckets
}

type ForEachShardCallback func(ctx context.Context, sh shard.Shard) error

func (s *ShardCluster) ForEachShard(ctx context.Context, callback ForEachShardCallback) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, sh := range s.shards {
		eg.Go(func() error {
			return callback(ctx, sh)
		})
	}
	err := eg.Wait()
	if err != nil {
		return fmt.Errorf("can't execute for each shard callback: %w", err)
	}

	return nil
}

type ForEachBucketCallback func(ctx context.Context, bucketPool *bucket.Pool) error

func (s *ShardCluster) ForEachBucket(ctx context.Context, callback ForEachBucketCallback) error {
	eg, ctx := errgroup.WithContext(ctx)

	for buck, sh := range s.bucketToShard {
		eg.Go(func() error {
			return callback(ctx, bucket.NewPool(buck, sh.Pool))
		})
	}
	err := eg.Wait()
	if err != nil {
		return fmt.Errorf("can't execute for each shard callback: %w", err)
	}

	return nil
}

func (s *ShardCluster) GetBucket(key string) bucket.Bucket {
	return s.keyToBucket(key)
}

func (s *ShardCluster) PickPool(key string) (*bucket.Pool, error) {
	buck := s.keyToBucket(key)
	sh, ok := s.bucketToShard[buck]
	if !ok {
		return nil, ErrNoSuchShard
	}

	return bucket.NewPool(buck, sh.Pool), nil
}

func (s *ShardCluster) PickPoolByBucket(buck bucket.Bucket) (*bucket.Pool, error) {
	sh, ok := s.bucketToShard[buck]
	if !ok {
		return nil, ErrNoSuchShard
	}

	return bucket.NewPool(buck, sh.Pool), nil
}

func (s *ShardCluster) GetShards() []shard.Shard {
	cp := make([]shard.Shard, len(s.shards))
	copy(cp, s.shards)
	return cp
}

func (s *ShardCluster) Close() {
	for _, sh := range s.shards {
		sh.Pool.Close()
	}
}
