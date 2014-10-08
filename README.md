Shell-like Configuration Files
==============================

A configuration file parser for Go, with syntax based on POSIX
`/bin/sh` word splitting, escaping, and expansion rules.

See here in [README.md](https://github.com/3ofcoins/shlike/) for config
syntax description.

See [GoDoc](https://godoc.org/pkg/github.com/3ofcoins/shlike/) for Go
API description.

Why new format?
---------------

Because there is no non-trivial, human-friendly configuration file
format. JSON is insufficient. YAML is overcomplicated and its
abstraction handling does not really work. Another formats are
obscure, idiosyncratic, as incomplete as JSON and little less
overcomplicated than YAML. Shell syntax and quoting rules are well
known, and expressive enough to allow complex configuration.

Configuration Format
--------------------

A configuration file evaluates to a sequence of lines. Each line is
a sequence of words. For example, configuration file:

  foo bar
  baz quux

has two lines, each of which contains two words.

The rules for escaping, expansion, and interpretation of special
characters resembles
[POSIX `/bin/sh` syntax](http://pubs.opengroup.org/onlinepubs/009604599/utilities/xcu_chap02.html):

1. One or more white space characters separate words
2. One or more line breaks separate lines. Any white space at
   beginning or end of line is discarded.
3. Any non-whitespace non-linebreak printable characters constitute
   words, with exception of: `$ \ " ' #`
4. An unquoted hash character (`#`) starts a comment. Any characters
   between `#` and end of line will be discarded. A comment will
   always introduce a line break, it is not possible to continue line
   that includes a comment.
5. An unquoted backslash will quote the following character as a
   literal word element, unless the following character is a line
   break.
6. An unquoted backslash at the end of the line (followed by a line
   break) continues the line: it removes backslash itself and the line
   break from the input.
7. Single quotes (`'`) around a sequence of characters will literally
   quote the enclosed characters, without any interpretation or
   escaping, as word elements. A single quote cannot occur within a
   single-quoted word.
8. Double quotes (`"`) around a sequence of characters will quote the
   enclosed characters, with the following exceptions:
   - Backslash escapes are interpreted as above (backslash removes
     line break, escapes any other character)
   - Variable expansion is performed (see below), but does not cause
     any word breaks. Variable value's words are inserted literally
     into the double-quoted word, joined with spaces or a specified
     string. To specify a glue string, enclose variable name in braces
     and use the `|` character after variable name. E.g. if `FOO`
     variable's value is `a b c` (three words), then `"${FOO|.}"` will
     expand to `a.b.c`; `"${FOO}"` will expand to `'a b c'`;
     `"${FOO|}"` will expand to `abc`.
9. The dollar sign (`$`) will expand variable. Variable name can be
   provided directly (`$FOO`) or enclosed in curly braces
   (`${FOO}`). As shlike's variable values are always _word
   sequences_, to avoid escaping issues and ambiguous situation, a
   variable expansion outside a double-quoted string will always end
   word, insert variable value's words separately into the output, and
   a new word will be started afterwards. E.g. if `FOO` variable's
   value is `a b c` (three words), then `1${FOO}2` will expand to `1 a
   b c 2` (five words). Within curly braces, a glue string
   specification is valid, by meaningless (i.e. `${FOO}` is exactly
   equivalent to `${FOO|whatever}`).
10. A valid variable name consists of a letter or underscore (`a-zA-Z_`),
    optionally followed by more letters, digits, or underscore
    characters (regular expression would be `[a-zA-Z_][a-zA-Z0-9]*`)

Source Directive
----------------

A line that begins with a single dot (`.`), followed by whitespace,
and then by a single word, will _source_ the file named by that word
(read the named file before continuing interpretation of current
file). If the file name is a relative path, it will be expanded from
the current file's directory, rather than working directory of the
process.

Variable Assignments
--------------------

A line that begins with a valid variable name, followed by one of `=`,
`+=`, or `?=`, optionally surrounded by white space, will read the
rest of the line according to regular rules, but the resulting list of
words will be assigned to the named variable. The different assignment
types are:

 - `=` will always set the variable's value. If variable already has a
   value, it will be discarded
 - `+=` will append the words to the variable's value
 - `?=` will discard the words and keep the existing value, if the
   variable has been already set

Full Example
------------

    # This is a comment. It will be ignored.
    # First, we will set some variables.
    META = foo bar baz quux # a list of four words
    NUMBERS = 4 8 15
    NUMBERS += \
        16 \
        23 \
        42 # all three numbers are one continued line
    META ?= these words will be ignored, meta has a value already
    SENTENCE ?= 'Lorem ipsum dolor sit amet' # This one will be set,
                                             # though
    one two three?
    . path/to/included.conf # will load the named file at this point
    'Meta is:' $META
    "Numbers are: \"${NUMBERS|, }\""
    Words\ not\ separated' by whitespace '"are joined together."
    Not expanded: \$META "\${META}"
    'single quotes\ retain\
    backslashes and $character'
    "double\ quotes inter\
    prete them"
    Back\
    slash\ dis\
    cards\ line\ breaks

This file will expand to the following line sequence (in JSON syntax):

    [
     ["one", "two", "three?"],
     […whatever was read from included.conf…],
     ["Meta is:", "foo", "bar", "baz", "quux"],
     ["Numbers are: \"4, 8, 15, 16, 23, 42\""],
     ["Words not separated by whitespace are joined together"],
     ["Not", "expanded:", "$META", "${META}"],
     ["Single quotes\\ retain\\\nbackslashes and $character"],
     ["Double quotes interprete them"],
     ["Backslash discards line breaks"]
    ]
    
