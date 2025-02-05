package blogger

import "context"

type llama3OnlyBlogger struct {
	model string
}

var _ Blogger = &llama3OnlyBlogger{}

func (b *llama3OnlyBlogger) UpdateInputs(ctx context.Context) error {
	return nil
}

func (b *llama3OnlyBlogger) WriteBlog(ctx context.Context) error {
	return nil
}
