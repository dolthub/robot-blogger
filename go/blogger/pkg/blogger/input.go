package blogger

import (
	"os"
	"path/filepath"
	"strings"
)

type Input interface {
	ID() string
	Path() string
}

type markdownBlogPostInput struct {
	prefix string
	path   string
}

func NewMarkdownBlogPostInput(prefix, path string) Input {
	return &markdownBlogPostInput{
		prefix: prefix,
		path:   path,
	}
}

func (i *markdownBlogPostInput) ID() string {
	id := strings.TrimSuffix(filepath.Base(i.path), filepath.Ext(i.path))
	id = strings.TrimPrefix(id, i.prefix)
	return strings.ReplaceAll(id, " ", "_")
}

func (i *markdownBlogPostInput) Path() string {
	return i.path
}

func NewMarkdownBlogPostInputsFromDir(dir string) ([]Input, error) {
	inputs := make([]Input, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		inputs = append(inputs, NewMarkdownBlogPostInput(dir, path))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return inputs, nil
}
