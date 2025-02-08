package writer

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"github.com/tmc/langchaingo/textsplitter"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type markdownBlogPostInput struct {
	prefix   string
	path     string
	docIndex int
	content  []byte
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

func (i *markdownBlogPostInput) DocIndex() int {
	return i.docIndex
}

func (i *markdownBlogPostInput) contentBytes() ([]byte, error) {
	if i.content != nil {
		return i.content, nil
	}
	content, err := os.ReadFile(i.path)
	if err != nil {
		return nil, err
	}
	i.content = content
	return content, nil
}

func (i *markdownBlogPostInput) Content() string {
	if i.content != nil {
		return string(i.content)
	}
	content, err := i.contentBytes()
	if err != nil {
		return ""
	}
	i.content = content
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

func SplitMarkdownBlogPostIntoInputs(prefix, path string) ([]Input, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	splitter := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithChunkSize(1024), // default is 512
		// textsplitter.WithChunkOverlap(128), // default is 100
		//textsplitter.WithCodeBlocks(true),
		//textsplitter.WithHeadingHierarchy(true),
	)

	docs, err := textsplitter.CreateDocuments(splitter, []string{string(content)}, nil)
	if err != nil {
		return nil, err
	}

	inputs := make([]Input, 0)
	for i, doc := range docs {
		//fmt.Printf("path: %s, doc index: %d, doc content: %s\n", path, i, doc.PageContent)
		inputs = append(inputs, &markdownBlogPostInput{
			prefix:   prefix,
			path:     path,
			docIndex: i,
			content:  []byte(doc.PageContent),
		})
	}

	return inputs, nil

}

func NewMarkdownBlogPostInputsFromDir(dir string) ([]Input, error) {
	inputs := make([]Input, 0)
	files := make([]string, 0)
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
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)

	// todo: remove this
	files = files[:1]

	for _, file := range files {
		ins, err := SplitMarkdownBlogPostIntoInputs(dir, file)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, ins...)
	}

	return inputs, nil
}
