package dbs

import (
	"context"
)

type ServerName string

const (
	Postgres ServerName = "postgres"
	Mysql    ServerName = "mysql"
	Noop     ServerName = "noop"
)

type Rows interface {
	Err() error
	Next() bool
	Scan(dest ...any) error
}
type QueryFunc func(ctx context.Context, rows Rows) error

type DatabaseServer interface {
	Name() ServerName
	QueryContext(ctx context.Context, queryFunc QueryFunc, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
