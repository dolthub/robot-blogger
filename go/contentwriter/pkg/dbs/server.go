package dbs

import (
	"context"
	"database/sql"
)

const (
	Postgres = "postgres"
	Mysql    = "mysql"
	Noop     = "noop"
)

type QueryFunc func(ctx context.Context, rows *sql.Rows) error

type DatabaseServer interface {
	Name() string
	QueryContext(ctx context.Context, queryFunc QueryFunc, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
