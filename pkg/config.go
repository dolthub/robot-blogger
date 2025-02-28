package pkg

import "github.com/tmc/langchaingo/textsplitter"

type Config struct {
	Runner                          Runner
	Model                           Model
	StoreType                       StoreType
	Host                            string
	User                            string
	Password                        string
	Port                            int
	VectorDimensions                int
	StoreName                       string
	Splitter                        textsplitter.TextSplitter
	IncludeFileFunc                 func(path string) bool
	PreContentSystemPrompt          string
	PostContentSystemPromptTemplate string
	RefineContextSystemPrompt       string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithRunner(runner Runner) *Config {
	c.Runner = runner
	return c
}

func (c *Config) WithModel(model Model) *Config {
	c.Model = model
	return c
}

func (c *Config) WithStoreType(storeType StoreType) *Config {
	c.StoreType = storeType
	return c
}

func (c *Config) WithHost(host string) *Config {
	c.Host = host
	return c
}

func (c *Config) WithUser(user string) *Config {
	c.User = user
	return c
}

func (c *Config) WithPassword(password string) *Config {
	c.Password = password
	return c
}

func (c *Config) WithPort(port int) *Config {
	c.Port = port
	return c
}

func (c *Config) WithVectorDimensions(vectorDimensions int) *Config {
	c.VectorDimensions = vectorDimensions
	return c
}

func (c *Config) WithStoreName(storeName string) *Config {
	c.StoreName = storeName
	return c
}

func (c *Config) WithSplitter(splitter textsplitter.TextSplitter) *Config {
	c.Splitter = splitter
	return c
}

func (c *Config) WithIncludeFileFunc(includeFileFunc func(path string) bool) *Config {
	c.IncludeFileFunc = includeFileFunc
	return c
}

func (c *Config) WithPreContentSystemPrompt(preContentSystemPrompt string) *Config {
	c.PreContentSystemPrompt = preContentSystemPrompt
	return c
}

func (c *Config) WithPostContentSystemPromptTemplate(postContentSystemPromptTemplate string) *Config {
	c.PostContentSystemPromptTemplate = postContentSystemPromptTemplate
	return c
}

func (c *Config) WithRefineContextSystemPrompt(refineContextSystemPrompt string) *Config {
	c.RefineContextSystemPrompt = refineContextSystemPrompt
	return c
}
