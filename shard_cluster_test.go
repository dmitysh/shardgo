package shardgo

import (
	"context"
	"github.com/dmitysh/shardgo/bucket"
	"github.com/dmitysh/shardgo/shard"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"os"
	"testing"
	"time"
)

const (
	dbTimeout = time.Second * 2
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
	ctx, cancel := s.newContextWithDBTimeout()
	defer cancel()

	sc, err := NewShardCluster(ctx, shard.HashingKeyToBucket(20),
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

func (s *shardClusterSuite) newContextWithDBTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), dbTimeout)
}

func (s *shardClusterSuite) TearDownSuite() {
	s.sc.Close()
}

func (s *shardClusterSuite) TestAll() {
}
