package shlike

import "strings"

// Returns `str` escaped as a single configuration word
func Escape(str string) string {
	// Is it a safe bare word?
	if idx := rxText.FindStringIndex(str); idx != nil && idx[1] == len(str) {
		return str
	} else {
		return "'" + strings.Replace(str, "'", "'\\''", -1) + "'"
	}

}

// Returns a string containing `strs` as a line of escaped words
func EscapeLine(strs []string) string {
	estrs := make([]string, len(strs))
	for i, str := range strs {
		estrs[i] = Escape(str)
	}
	return strings.Join(estrs, " ")
}
