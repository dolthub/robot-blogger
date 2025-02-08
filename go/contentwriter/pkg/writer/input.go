package writer

type Input interface {
	ID() string
	Path() string
	DocIndex() int
	Content() string
	ContentMd5() (string, error)
}
