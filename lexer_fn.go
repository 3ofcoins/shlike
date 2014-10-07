package shlike

import "regexp"
import "unicode"

type lexFn func(*lexer) lexFn

var rxComment = regexp.MustCompile(`(?s)^#[^\n]*(\r?\n)*`)

func lexComment(l *lexer) lexFn {
	if l.matchStart(rxComment) == nil {
		l.errf("CAN'T HAPPEN in comment")
		return nil
	}
	return lexNewLine
}

func lexBackslash(l *lexer) lexFn {
	l.discard() // drop the actual backslash character
	switch r := l.next(); {
	case r == '\r' && l.peek() == '\n':
		l.next()
		fallthrough
	case r == '\n':
		if l.dquo {
			// escaped newline in double quotes is discarded
			l.discard()
		} else {
			l.emit(tkSpace)
		}
	default:
		l.emit(tkText)
	}
	return lexNewToken
}

var rxWhiteSpace = regexp.MustCompile(`^([\t\f ]|\\\r?\n)+`)

func lexWhiteSpace(l *lexer) lexFn {
	if l.matchStart(rxWhiteSpace) == nil {
		l.errf("Invalid white space")
		return nil
	}
	l.emit(tkSpace)
	return lexNewToken
}

var rxSingleQuoted = regexp.MustCompile(`(?s)^'([^']*)'`)

func lexSingleQuoted(l *lexer) lexFn {
	if pos := l.matchStart(rxSingleQuoted); pos == nil {
		l.errf("Invalid single quoted string")
		return nil
	} else {
		l.emitSubstring(tkText, pos[0], pos[1])
		return lexNewToken
	}
}

func lexDoubleQuote(l *lexer) lexFn {
	if !l.dquo && l.peek() == '"' {
		// Empty double quotes are a special case
		l.next()
		l.discard()
		l.addText("")
		return lexNewToken
	}
	// Just toggles inside-double-quotes status
	l.discard()
	l.dquo = !l.dquo
	return lexNewToken
}

var rxVariableReference = regexp.MustCompile(`^\$([-0-9#!$%&*+,.:;<=>?@^_/|~]|[_\pL][_\pL\pN]*|{((?s).*?)})`)

func lexVariableReference(l *lexer) lexFn {
	pos := l.matchStart(rxVariableReference)
	if pos == nil {
		l.errf("Invalid variable reference")
		return nil
	}
	if pos[2] < 0 {
		l.emitSubstring(tkVariableReference, pos[0], pos[1])
	} else {
		l.emitSubstring(tkVariableReference, pos[2], pos[3])
	}
	return lexNewToken
}

var rxText = regexp.MustCompile(`^[^\\'"$#[:space:]]+`)
var rxTextDquo = regexp.MustCompile(`^[^\\"$]+`)

func lexText(l *lexer) lexFn {
	var rx *regexp.Regexp
	if l.dquo {
		rx = rxTextDquo
	} else {
		rx = rxText
	}
	if l.matchStart(rx) == nil {
		l.errf("Invalid text?")
		return nil
	}
	l.emit(tkText)
	return lexNewToken
}

func lexNewToken(l *lexer) lexFn {
	r := l.next()
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
			return lexText
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

var rxLineBreak = regexp.MustCompile(`^(\r?\n)+`)

func lexLineBreak(l *lexer) lexFn {
	if l.matchStart(rxLineBreak) == nil {
		l.errf("Invalid newline, probably runaway line feed character.")
		return nil
	}
	return lexNewLine
}

var rxNewLineWhitespace = regexp.MustCompile(`^\s*`)
var rxNewLineAssignment = regexp.MustCompile(`^([_\pL][_\pL\pN]*)[\t\v\f ]*([:?+]?)=[\t\v\f ]*`)

func lexNewLine(l *lexer) lexFn {
	l.emit(tkEOL)
	l.matchStart(rxNewLineWhitespace)
	l.discard()
	if pos := l.matchStart(rxNewLineAssignment); pos != nil {
		l.assignTo = l.substring(pos[0], pos[1])
		l.assignBy = asgmtOverwrite
		switch l.substring(pos[2], pos[3]) {
		case "+":
			l.assignBy = asgmtAppend
		case "?":
			l.assignBy = asgmtKeep
		}
		l.discard()
	}
	return lexNewToken
}

func lexEOF(l *lexer) lexFn {
	l.emit(tkEOL)
	return nil
}
