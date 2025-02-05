package blogger

import (
	"io"
)

type Input interface {
	ID() string
	io.Reader
}
