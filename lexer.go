// Syntax
//
// Lexer implementation is heavily inspired by Go's own template lexer,
// as described in https://cuddle.googlecode.com/hg/talk/lex.html
package shlike

import "fmt"
import "os"
import "regexp"
import "strings"
import "unicode/utf8"

type tkKind int

const (
	tkText tkKind = iota
	tkSpace
	tkEOL
	tkVariableReference
)

type opKind int

const (
	opLine opKind = iota
	opSet
	opAppend
	opSetIfUnset
)

type op struct {
	kind opKind
	name string
}

const eof = -1

type lexer struct {
	*Config
	name                  string
	data                  string
	start, pos, width, ln int
	welt, line            []string
	dquo                  bool
	err                   error
	opName                string
	op                    opKind
}

func (l *lexer) parse() error {
	for state := lexBOL; state != nil; {
		state = state(l)
	}
	return l.err
}

func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.data[:l.start], "\n")
}

func (l *lexer) debugPrefix(start, pos int) string {
	ln := 1 + strings.Count(l.data[:pos], "\n")
	lpos := pos - strings.LastIndex(l.data[:pos], "\n")
	before := start - 3
	after := pos + 3
	if before < 0 {
		before = 0
	}
	if after > len(l.data) {
		after = len(l.data)
	}
	return fmt.Sprintf("%s:%d:%d:\t%#v.%#v.%#v", l.name, ln, lpos, l.data[before:start], l.data[start:pos], l.data[pos:after])
}

func (l *lexer) debug(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\t%s\n", l.debugPrefix(l.start, l.pos), fmt.Sprintf(format, v...))
}

func (l *lexer) addText(text string) {
	if l.ln == 0 {
		l.ln = l.lineNumber()
	}
	l.welt = append(l.welt, text)
}

func (l *lexer) endWord() {
	if len(l.welt) > 0 {
		l.line = append(l.line, strings.Join(l.welt, ""))
	}
	l.welt = nil
}

func (l *lexer) expandReference(vref string) {
	val := l.Get(vref)
	if val == nil {
		l.warnf("Undefined variable %#v", vref)
		return
	}
	if l.dquo {
		l.welt = append(l.welt, strings.Join(val, " "))
	} else {
		l.endWord()
		l.line = append(l.line, val...)
	}
}

func (l *lexer) endLine() {
	l.endWord()
	switch l.op {
	case opLine:
		if len(l.line) > 0 {
			l.addLine(l.line)
		}
	case opSet:
		l.Set(l.opName, l.line...)
	case opAppend:
		l.Append(l.opName, l.line...)
	case opSetIfUnset:
		if l.Get(l.opName) == nil {
			l.Set(l.opName, l.line...)
		}
	}
	l.opName = ""
	l.op = opLine
	l.line = nil
	l.ln = 0
}

func (l *lexer) errf(format string, args ...interface{}) {
	l.err = fmt.Errorf("%s: %s", l.debugPrefix(l.start, l.pos), fmt.Sprintf(format, args...))
}

func (l *lexer) warnf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: WARNING: %s\n", l.debugPrefix(l.start, l.pos), fmt.Sprintf(format, args...))
}

func (l *lexer) decodeNextRune() (rune, int) {
	if l.pos >= len(l.data) {
		return eof, 0
	}
	return utf8.DecodeRuneInString(l.data[l.pos:])
}

func (l *lexer) next() rune {
	r, w := l.decodeNextRune()
	l.pos += w
	l.width = w
	return r
}

// Rewinds `pos` back to `start`
func (l *lexer) rew() {
	l.pos = l.start
	l.width = 0
}

func (l *lexer) peek() rune {
	r, _ := l.decodeNextRune()
	return r
}

// Tries to match `rx` against current position. `rx` has to be
// start-anchored (beginning with `^`). Returns nil if not found,
// array of submatch positions (empty array if no submatches)
// otherwise. If `rx` matched, current position is pushed forward by
// length of the match, and width is set to match length (so that
// `l.back()` will go to beginning of match)
func (l *lexer) match(rx *regexp.Regexp) []int {
	if loc := rx.FindStringSubmatchIndex(l.data[l.pos:]); loc == nil {
		return nil
	} else {
		l.pos += loc[1]
		l.width = loc[1]
		return loc[2:]
	}
}

func (l *lexer) discard() {
	l.start = l.pos
	l.width = 0
}

func (l *lexer) consume() (region string) {
	region = l.data[l.start:l.pos]
	l.discard()
	return
}
