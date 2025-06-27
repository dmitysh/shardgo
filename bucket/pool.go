package bucket

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	bucketSchemaPlaceholder = "_bucket_"
	schemaPattern           = "bucket_%s"
)

// Pool connection pool for bucket's schema
type Pool struct {
	bucketName string
	pool       *pgxpool.Pool
}

// NewPool Pool constructor
func NewPool(bucket Bucket, pool *pgxpool.Pool) *Pool {
	return &Pool{
		pool:       pool,
		bucketName: bucket.String(),
	}
}

// Exec pgx Exec
func (p *Pool) Exec(ctx context.Context, q string, args ...interface{}) (pgconn.CommandTag, error) {
	return p.pool.Exec(ctx, p.replaceSchema(q), args...)
}

// Query pgx Query
func (p *Pool) Query(ctx context.Context, q string, args ...interface{}) (pgx.Rows, error) {
	return p.pool.Query(ctx, p.replaceSchema(q), args...)
}

// QueryRow pgx QueryRow
func (p *Pool) QueryRow(ctx context.Context, q string, args ...interface{}) pgx.Row {
	return p.pool.QueryRow(ctx, p.replaceSchema(q), args...)
}

func (p *Pool) replaceSchema(q string) string {
	return strings.ReplaceAll(q, bucketSchemaPlaceholder, fmt.Sprintf(schemaPattern, p.bucketName))
}
