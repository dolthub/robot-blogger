package blogger

import (
	"context"
)

type Blogger interface {
	UpdateInput(ctx context.Context, input Input) error
	WriteBlog(ctx context.Context) error
}
