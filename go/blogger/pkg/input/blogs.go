package input

import (
	"time"
)

type blogInput struct {
}

var _ Input = &blogInput{}

func (b *blogInput) ID() string {
	return ""
}

func (b *blogInput) Content() string {
	return ""
}

func (b *blogInput) CreatedAt() time.Time {
	return time.Time{}
}
