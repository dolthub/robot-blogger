package pkg

import "github.com/tmc/langchaingo/textsplitter"

type noopTextSplitter struct{}

var _ textsplitter.TextSplitter = (*noopTextSplitter)(nil)

func (ns noopTextSplitter) SplitText(text string) ([]string, error) {
	panic("unimplemented")
}

func NewNoopTextSplitter() *noopTextSplitter {
	return &noopTextSplitter{}
}
