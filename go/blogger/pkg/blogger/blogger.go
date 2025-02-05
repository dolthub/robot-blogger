package blogger

import (
	"context"
)

type Blogger interface {
	UpdateInputs(ctx context.Context) error
	WriteBlog(ctx context.Context) error
}
