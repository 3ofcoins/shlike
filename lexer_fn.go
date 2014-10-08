package shlike

import "regexp"
import "unicode"

type lexFn func(*lexer) lexFn

func lexByRx(name string, rx *regexp.Regexp, inner func(*lexer, string, []int) lexFn) lexFn {
	return func(l *lexer) lexFn {
		l.rew()
		if pos := l.match(rx); pos == nil {
			l.errf("Invalid %s", name)
			return nil
		} else {
			return inner(l, l.consume(), pos)
		}
	}
}

var lexBackslash lexFn
var rxBackslash = regexp.MustCompile(`^\\(\r?\n|.)`)

func lexBackslashByRx(l *lexer, region string, pos []int) lexFn {
	val := region[pos[0]:pos[1]]
	if val[len(val)-1] == '\n' {
		// Double-quoted escaped newline is discarded (interpreted as
		// empty string). Otherwise, it's a white space (word separator).
		if !l.dquo {
			l.endWord()
		}
	} else {
		l.addText(val)
	}
	return lexDispatch
}

var lexWhiteSpace lexFn
var rxWhiteSpace = regexp.MustCompile(`^([\t\f ]|\\\r?\n)+`)

func lexWhiteSpaceByRx(l *lexer, _ string, _ []int) lexFn {
	l.endWord()
	return lexDispatch
}

var lexSingleQuoted lexFn
var rxSingleQuoted = regexp.MustCompile(`(?s)^'([^']*)'`)

func lexSingleQuotedByRx(l *lexer, region string, pos []int) lexFn {
	l.addText(region[pos[0]:pos[1]])
	return lexDispatch
}

func lexDoubleQuote(l *lexer) lexFn {
	l.next()
	if !l.dquo && l.peek() == '"' {
		// Empty double quotes shortcut
		l.next()
		l.discard()
		l.addText("")
		return lexDispatch
	}
	// Just toggles inside-double-quotes status
	l.discard()
	l.dquo = !l.dquo
	return lexDispatch
}

var lexVariableReference lexFn
var rxVariableReference = regexp.MustCompile(`^\$([-0-9#!$%&*+,.:;<=>?@^_/|~]|[_\pL][_\pL\pN]*|{((?s).*?)})`)

func lexVariableReferenceByRx(l *lexer, region string, pos []int) lexFn {
	if pos[2] < 0 {
		l.expandReference(region[pos[0]:pos[1]])
	} else {
		l.expandReference(region[pos[2]:pos[3]])
	}
	return lexDispatch
}

var lexText, lexTextDquo lexFn
var rxText = regexp.MustCompile(`^[^\\'"$#[:space:]]+`)
var rxTextDquo = regexp.MustCompile(`^[^\\"$]+`)

func lexTextByRx(l *lexer, region string, _ []int) lexFn {
	l.addText(region)
	return lexDispatch
}

var lexLineBreak, lexComment lexFn
var rxLineBreak = regexp.MustCompile(`^(\r?\n)+`)
var rxComment = regexp.MustCompile(`(?s)^#[^\n]*(?:\r?\n)*`)

func lexEOLByRx(l *lexer, _ string, _ []int) lexFn {
	l.endLine()
	return lexBOL
}

var lexBOL lexFn
var rxBOL = regexp.MustCompile(`^\s*(?:([_\pL][_\pL\pN]*)[\t\v\f ]*([?+]?)=[\t\v\f ]*)?`)

func lexBOLByRx(l *lexer, region string, pos []int) lexFn {
	if pos[0] >= 0 {
		// Assignment
		l.opName = region[pos[0]:pos[1]]
		l.op = opSet
		switch region[pos[2]:pos[3]] {
		case "+":
			l.op = opAppend
		case "?":
			l.op = opSetIfUnset
		}
	}
	return lexDispatch
}

func lexDispatch(l *lexer) lexFn {
	r := l.peek()
	if l.dquo {
		// inside double-quoted string
		switch r {
		case '\\':
			return lexBackslash
		case '$':
			return lexVariableReference
		case '"':
			return lexDoubleQuote
		case eof:
			l.errf("Unclosed double quoted string")
			return nil
		default:
			return lexTextDquo
		}
	} else {
		// not inside double-quoted string
		switch r {
		case '\r', '\n':
			return lexLineBreak
		case '\\':
			return lexBackslash
		case '#':
			return lexComment
		case '$':
			return lexVariableReference
		case '\'':
			return lexSingleQuoted
		case '"':
			return lexDoubleQuote
		case eof:
			return lexEOF
		default:
			if unicode.IsSpace(r) {
				return lexWhiteSpace
			} else {
				return lexText
			}
		}
	}
}

func lexEOF(l *lexer) lexFn {
	l.discard()
	l.endLine()
	return nil
}

func init() {
	lexBackslash = lexByRx("backslash escape", rxBackslash, lexBackslashByRx)
	lexWhiteSpace = lexByRx("whitespace", rxWhiteSpace, lexWhiteSpaceByRx)
	lexSingleQuoted = lexByRx("single quoted string", rxSingleQuoted, lexSingleQuotedByRx)
	lexVariableReference = lexByRx("variable reference", rxVariableReference, lexVariableReferenceByRx)
	lexText = lexByRx("bare text", rxText, lexTextByRx)
	lexTextDquo = lexByRx("double quoted text", rxTextDquo, lexTextByRx)
	lexLineBreak = lexByRx("line break", rxLineBreak, lexEOLByRx)
	lexComment = lexByRx("comment", rxComment, lexEOLByRx)
	lexBOL = lexByRx("new line", rxBOL, lexBOLByRx)
}
