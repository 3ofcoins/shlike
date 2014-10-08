package shlike

import "fmt"
import "io/ioutil"
import "sort"
import "strings"

// A configuration object. Can load multiple files, evaluate
// configuration strings, hold variables, and execute dot-directives
// during evaluation.
type Config struct {
	variables map[string][]string
	lines     [][]string
	dot       map[string]DotCommand
}

// Returns new empty config object
func NewConfig() *Config {
	rv := &Config{map[string][]string{}, [][]string{}, map[string]DotCommand{}}
	for name, handler := range DotCommands {
		rv.dot[name] = handler
	}
	return rv
}

// Returns contents as string, escaped as a loadable config, fully
// expanded and stripped of comments.
//
// BUG: if a single-character variable has been set from code, it will
// be serialized, but won't be loaded correctly.
func (c *Config) String() string {
	pieces := make([]string, 0, len(c.variables)+len(c.lines))

	for n, v := range c.variables {
		pieces = append(pieces, fmt.Sprintf("%s = %s", n, EscapeLine(v)))
	}

	// Sort variables alphabetically
	sort.Strings(pieces)

	for _, ln := range c.lines {
		pieces = append(pieces, EscapeLine(ln))
	}
	return strings.Join(pieces, "\n")
}

func (c *Config) Set(name string, values ...string) {
	if values == nil {
		values = []string{}
	}
	c.variables[name] = values
}

func (c *Config) Append(name string, values ...string) {
	c.variables[name] = append(c.variables[name], values...)
}

func (c *Config) Get(name string) []string {
	return c.variables[name]
}

func (c *Config) Unset(name string) {
	delete(c.variables, name)
}

func (c *Config) addLine(words []string) {
	c.lines = append(c.lines, words)
}

func (c *Config) lexer(name, data string) *lexer {
	return &lexer{Config: c, name: name, data: data}
}

func (c *Config) eval(name, data string) error {
	return c.lexer(name, data).parse()
}

func (c *Config) Eval(config string) error {
	return c.eval("(eval)", config)
}

func (c *Config) Load(path string) error {
	config, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return c.eval(path, string(config))
}

func (c *Config) Dot(name string, cmd DotCommand) {
	if cmd == nil {
		delete(c.dot, name)
	} else {
		c.dot[name] = cmd
	}
}
