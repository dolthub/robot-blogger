package blogger

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Input interface {
	ID() string
	Path() string
	Content() string
	ContentMd5() (string, error)
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

func (i *markdownBlogPostInput) contentBytes() ([]byte, error) {
	return os.ReadFile(i.path)
}

func (i *markdownBlogPostInput) Content() string {
	content, err := i.contentBytes()
	if err != nil {
		return ""
	}
	return string(content)
}

func (i *markdownBlogPostInput) ContentMd5() (string, error) {
	content, err := i.contentBytes()
	if err != nil {
		return "", err
	}
	r := bytes.NewReader(content)
	hash := md5.New()
	_, err = io.Copy(hash, r)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
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
