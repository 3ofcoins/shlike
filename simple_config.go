package shlike

import "io/ioutil"

// An implementation of `Config` interface.
type SimpleConfig struct {
	Vars  map[string][]string // Variable values
	Lines [][]string          // Evaluated lines
}

// Returns new config object
func NewConfig() *SimpleConfig {
	return &SimpleConfig{map[string][]string{}, [][]string{}}
}

func (c *SimpleConfig) ReceiveLine(words []string) {
	c.Lines = append(c.Lines, words)
}

func (c *SimpleConfig) Set(variable string, values ...string) {
	if values == nil {
		values = []string{}
	}
	c.Vars[variable] = values
}

func (c *SimpleConfig) Append(variable string, values ...string) {
	c.Vars[variable] = append(c.Vars[variable], values...)
}

func (c *SimpleConfig) Get(variable string) []string {
	return c.Vars[variable]
}

func (c *SimpleConfig) Unset(variable string) {
	delete(c.Vars, variable)
}

func (c *SimpleConfig) Variables() []string {
	rv := make([]string, 0, len(c.Vars))
	for name, _ := range c.Vars {
		rv = append(rv, name)
	}
	return rv
}

func (c *SimpleConfig) Length() int {
	return len(c.Lines)
}

func (c *SimpleConfig) Line(number int) []string {
	if number < 0 || number >= len(c.Lines) {
		return nil
	}
	return c.Lines[number]
}

func (c *SimpleConfig) Iter() <-chan []string {
	ch := make(chan []string)
	go func() {
		for _, ln := range c.Lines {
			ch <- ln
		}
		close(ch)
	}()
	return ch
}

// Evaluates `source` configuration string
func (c *SimpleConfig) Eval(source string) error {
	return EvalInto(c, source)
}

// Loads configuration from `path`
func (c *SimpleConfig) Load(path string) error {
	return LoadInto(c, path)
}

// Serializes configuration into a loadable string
func (c *SimpleConfig) Serialize() string {
	return Serialize(c)
}

// Saves configuration into a file
func (c *SimpleConfig) Save(path string) error {
	return ioutil.WriteFile(path, []byte(c.Serialize()+"\n"), 0666)
}
