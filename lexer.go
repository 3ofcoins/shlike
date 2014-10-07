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

type asgmtKind int

const (
	asgmtNone asgmtKind = iota
	asgmtOverwrite
	asgmtAppend
	asgmtKeep
)

type asgmt struct {
	kind asgmtKind
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
	assignTo              string
	assignBy              asgmtKind
}

func newLexer(cfg *Config, name, data string) *lexer {
	return &lexer{Config: cfg, name: name, data: data}
}

func (l *lexer) parse() error {
	for state := lexNewLine; state != nil; {
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
	if l.assignBy != asgmtNone {
		if l.Get(l.assignTo) != nil {
			switch l.assignBy {
			case asgmtOverwrite:
				l.Set(l.assignTo, l.line...)
			case asgmtAppend:
				l.Append(l.assignTo, l.line...)
			case asgmtKeep:
				// discard values
			}
		} else {
			l.Set(l.assignTo, l.line...)
		}
		l.line = nil
		l.ln = 0
	} else {
		if len(l.line) > 0 {
			l.addLine(l.line)
			l.line = nil
			l.ln = 0
		}
	}
	l.assignTo = ""
	l.assignBy = asgmtNone
}

func (l *lexer) doEmit(tk tkKind, start, pos int, text string) {
	if text == "" {
		text = l.data[start:pos]
	}
	// fmt.Printf("%s:\t%d\t%#v\n", l.debugPrefix(start, pos), tk, text)
	switch tk {
	case tkText:
		l.addText(text)
	case tkVariableReference:
		l.expandReference(text)
	case tkSpace:
		l.endWord()
	case tkEOL:
		l.endLine()
	default:
		fmt.Fprintf(os.Stderr, "%s: Unknown token %d: %#v\n", l.debugPrefix(start, pos), tk, text)
	}
}

func (l *lexer) errf(format string, args ...interface{}) {
	l.err = fmt.Errorf("%s: %s", l.debugPrefix(l.start, l.pos), fmt.Sprintf(format, args...))
}

func (l *lexer) warnf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: WARNING: %s\n", l.debugPrefix(l.start, l.pos), fmt.Sprintf(format, args...))
}

func (l *lexer) substring(start, end int) string {
	regw := l.pos - l.start
	if start < 0 {
		start = 0
	}
	if start > regw {
		start = regw
	}
	if end < 0 || end > regw {
		end = regw
	}
	return l.data[l.start+start : l.start+end]
}

func (l *lexer) emit(tk tkKind) {
	l.doEmit(tk, l.start, l.pos, "")
	l.start = l.pos
}

func (l *lexer) emitSubstring(tk tkKind, start, end int) {
	regw := l.pos - l.start
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = regw
	}
	if start > regw || end > regw {
		l.errf("Substring %d:%d outside %d chars wide region %d:%d", start, end, regw, l.start, l.pos)
	}
	l.doEmit(tk, l.start+start, l.start+end, "")
	l.start = l.pos
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

func (l *lexer) back() {
	if l.pos -= l.width; l.pos < 0 {
		l.pos = 0
	}
}

func (l *lexer) forth() {
	l.pos += l.width
}

// Fast-forwards position by `length`
func (l *lexer) ff(length int) {
	l.pos += length
	l.width = length
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

// Try to match at start of current token rather than current
// position. This is same as calling l.rew() before l.match(), but it
// will keep pos/width if not matched.
func (l *lexer) matchStart(rx *regexp.Regexp) []int {
	savedPos := l.pos
	savedWidth := l.width
	l.rew()
	if rv := l.match(rx); rv == nil {
		l.pos = savedPos
		l.width = savedWidth
		return nil
	} else {
		return rv
	}
}

func (l *lexer) discard() {
	l.start = l.pos
}
