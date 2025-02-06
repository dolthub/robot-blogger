package models

import (
	"context"
	"io"

	"github.com/dolthub/robot-blogger/go/blogger/pkg/dbs"
)

type Chatter interface {
	Chat(ctx context.Context, input string, wc io.WriteCloser) (int64, error)
	ChatWithEmbeddings(ctx context.Context, input string, db dbs.DatabaseServer, wc io.WriteCloser) (int64, error)
}
