package shardgo

import (
	"context"
	"os"
	"sync/atomic"
	"testing"

	"github.com/dmitysh/shardgo/bucket"
	"github.com/dmitysh/shardgo/shard"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

const (
	bucketCount = 6
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestSuiteShardCluster(t *testing.T) {
	t.Parallel()

	s := newShardClusterSuite()
	suite.Run(t, s)
}

type shardClusterSuite struct {
	suite.Suite

	sc *ShardCluster
}

func newShardClusterSuite() *shardClusterSuite {
	return &shardClusterSuite{}
}

func (s *shardClusterSuite) SetupSuite() {
	s.sc = s.newTestShardCluster()
}

func (s *shardClusterSuite) newTestShardCluster() *ShardCluster {
	ctx := s.T().Context()

	sc, err := NewShardCluster(ctx, shard.HashingKeyToBucket(bucketCount),
		[]shard.BucketToShard{
			{
				ShardDSN: shard.DSN(os.Getenv("SHARD_1_DSN")),
				BucketsList: []bucket.Range{
					{
						From: 1,
						To:   2,
					},
				},
			},
			{
				ShardDSN: shard.DSN(os.Getenv("SHARD_2_DSN")),
				BucketsList: []bucket.Range{
					{
						From: 3,
						To:   4,
					},
				},
			},
			{
				ShardDSN: shard.DSN(os.Getenv("SHARD_3_DSN")),
				BucketsList: []bucket.Range{
					{
						From: 5,
						To:   6,
					},
				},
			},
		})
	s.Require().NoError(err)

	return sc
}

func (s *shardClusterSuite) TearDownSuite() {
	s.sc.Close()
}

func (s *shardClusterSuite) SetupTest() {
	s.clearUserTable()
}

func (s *shardClusterSuite) clearUserTable() {
	ctx := s.T().Context()
	err := s.sc.ForAllBuckets(ctx, func(ctx context.Context, bucketPool *bucket.Pool) error {
		q := `TRUNCATE TABLE _bucket_.user`
		_, err := bucketPool.Exec(ctx, q)
		if err != nil {
			return err
		}

		return nil
	})
	s.Require().NoError(err)
}

func (s *shardClusterSuite) TestForAllShards() {
	ctx := s.T().Context()
	want := map[string][]string{
		os.Getenv("SHARD_1_DSN"): {"bucket_1", "bucket_2"},
		os.Getenv("SHARD_2_DSN"): {"bucket_3", "bucket_4"},
		os.Getenv("SHARD_3_DSN"): {"bucket_5", "bucket_6"},
	}

	var shardsCnt atomic.Int32
	err := s.sc.ForAllShards(ctx, func(ctx context.Context, sh shard.Shard) error {
		q := `SELECT schema_name
				FROM information_schema.schemata
			   WHERE schema_name NOT LIKE 'pg_%' 
				 AND schema_name NOT IN ('information_schema', 'public')
`
		rows, err := sh.Pool.Query(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()

		schemas := make([]string, 0)
		for rows.Next() {
			var schemaName string
			err = rows.Scan(&schemaName)
			if err != nil {
				return err
			}
			schemas = append(schemas, schemaName)
		}

		s.Require().ElementsMatch(want[string(sh.DSN)], schemas)
		shardsCnt.Add(1)

		return nil
	})
	s.Require().NoError(err)
	s.Require().Len(want, int(shardsCnt.Load()))
}

func (s *shardClusterSuite) TestForAllBuckets() {
	ctx := s.T().Context()

	var bucketsCnt atomic.Int32
	err := s.sc.ForAllBuckets(ctx, func(ctx context.Context, bucketPool *bucket.Pool) error {
		q := `INSERT INTO _bucket_.user
			  VALUES ($1, $2)`
		_, err := bucketPool.Exec(ctx, q, uuid.NewString(), "Vasya")
		if err != nil {
			return err
		}
		bucketsCnt.Add(1)

		return nil
	})
	s.Require().NoError(err)

	var usersTotal atomic.Int32
	err = s.sc.ForAllBuckets(ctx, func(ctx context.Context, bucketPool *bucket.Pool) error {
		q := `SELECT * FROM _bucket_.user`
		rows, err := bucketPool.Query(ctx, q) // nolint: govet
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id, name string
			err = rows.Scan(&id, &name)
			if err != nil {
				return err
			}
			s.Require().Equal("Vasya", name)
		}
		usersTotal.Add(1)

		return nil
	})
	s.Require().NoError(err)
	s.Require().Equal(bucketsCnt.Load(), usersTotal.Load())
}

func (s *shardClusterSuite) TestPickPool() {
	ctx := s.T().Context()

	// All keys distributed for different buckets
	keys := []string{"abc", "1", "keks"}
	buckets := make([]bucket.Bucket, 0)
	for _, key := range keys {
		pool, err := s.sc.PickPool(key)
		s.Require().NoError(err)

		q := `INSERT INTO _bucket_.user 
			  VALUES ($1, $2)`
		_, err = pool.Exec(ctx, q, uuid.NewString(), "Spitz")
		s.Require().NoError(err)

		buckets = append(buckets, s.sc.GetBucket(key))
	}

	for _, buck := range buckets {
		pool, err := s.sc.PickPoolByBucket(buck)
		s.Require().NoError(err)

		q := `SELECT user_id, name FROM _bucket_.user`
		row := pool.QueryRow(ctx, q)

		var id, name string
		err = row.Scan(&id, &name)
		s.Require().NoError(err)

		s.Require().Equal("Spitz", name)
	}
}
