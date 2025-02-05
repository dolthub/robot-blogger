package models

import "context"

type Chatter interface {
	Chat(ctx context.Context, input string) (string, error)
}
