package models

import "context"

type ModelServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
