// `shlike` implements a flexible configuration file format, based on
// POSIX `/bin/sh` word splitting, quoting, and expansion. See the
// README.md file (https://github.com/3ofcoins/shlike/) for description
// of the configuration syntax, and GoDoc
// (https://godoc.org/pkg/github.com/3ofcoins/shlike/) for API details.
package shlike

import "fmt"
import "io/ioutil"
import "sort"
import "strings"

// A configuration object interface. Can accept variables,
// dot-commands, and parsed lines from lexer, and exposes them to end
// user.
type Config interface {
	ReceiveLine(words []string)           // Receive a config line from lexer
	Set(name string, values ...string)    // Set variable
	Append(name string, values ...string) // Append to variable
	Get(name string) []string             // Get variable's value
	Unset(name string)                    // Unset variable
	Variables() []string                  // List of variable names
	Length() int                          // Number of config lines
	Line(number int) []string             // A single line
	Iter() <-chan []string                // Iterator over lines
}

// Evaluates configuration string `source` into `cfg`. Wrapped by Convenience.Eval()
func EvalInto(cfg Config, source string) error {
	return newLexer(cfg, "(eval)", source).parse()
}

// Loads configuration file at `path` into `cfg`. Wrapped by Convenience.Load()
func LoadInto(cfg Config, path string) error {
	config, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return newLexer(cfg, path, string(config)).parse()
}

// Returns config serialized as an `EvalInto`-able configuration
// string. Wrapped by Convenience.Serialize()
func Serialize(c Config) string {
	vars := c.Variables()
	sort.Strings(vars)

	pieces := make([]string, 0, len(vars)+c.Length())

	for _, v := range vars {
		pieces = append(pieces, fmt.Sprintf("%s = %s", v, EscapeLine(c.Get(v))))
	}

	for ln := range c.Iter() {
		pieces = append(pieces, EscapeLine(ln))
	}
	return strings.Join(pieces, "\n")
}
