package input

import "time"

type Input interface {
	ID() string
	Content() string
	CreatedAt() time.Time
}
