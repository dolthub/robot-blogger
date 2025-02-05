package blogger

import (
	"io"
)

type Input interface {
	io.Reader
}
