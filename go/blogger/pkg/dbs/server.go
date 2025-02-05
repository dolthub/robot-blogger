package dbs

import (
	"context"
)

type DatabaseServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetConnectionString() string
}
