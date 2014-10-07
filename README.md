# shlike
--
    import "github.com/3ofcoins/shlike"

### Shell-like Configuration

Why new format? Because there is no non-trivial, human-friendly configuration
file format. JSON is insufficient. YAML is overcomplicated and its abstraction
handling is just as insufficient. Another formats are obscure, idiosyncratic, as
incomplete as JSON and little less overcomplicated than YAML. And Shell syntax
and quoting rules is well known.


### Syntax

Lexer implementation is heavily inspired by Go's own template lexer, as
described in https://cuddle.googlecode.com/hg/talk/lex.html

## Usage

#### func  Escape

```go
func Escape(str string) string
```
Returns `str` escaped as a single configuration word

#### func  EscapeLine

```go
func EscapeLine(strs []string) string
```
Returns a string containing `strs` as a line of escaped words

#### type Config

```go
type Config struct {
}
```

A configuration object. Can load multiple files, evaluate configuration strings,
hold variables, and execute dot-directives during evaluation.

#### func  NewConfig

```go
func NewConfig() *Config
```
Returns new empty config object

#### func (*Config) Append

```go
func (c *Config) Append(name string, values ...string)
```

#### func (*Config) Eval

```go
func (c *Config) Eval(config string) error
```

#### func (*Config) Get

```go
func (c *Config) Get(name string) []string
```

#### func (*Config) Load

```go
func (c *Config) Load(path string) error
```

#### func (*Config) Set

```go
func (c *Config) Set(name string, values ...string)
```

#### func (*Config) String

```go
func (c *Config) String() string
```
Returns contents as string, escaped as a loadable config, fully expanded and
stripped of comments.

BUG: if a single-character variable has been set from code, it will be
serialized, but won't be loaded correctly.

#### func (*Config) Unset

```go
func (c *Config) Unset(name string)
```
