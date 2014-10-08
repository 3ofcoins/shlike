package shlike

import "fmt"
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

func snakeToCamel(str string) string {
	return strings.Replace(
		strings.Title(
			strings.Replace(strings.ToLower(str),
				"_", " ", -1)),
		" ", "", -1)
}

func recoverFrom(wrapped func() error) error {
	ch := make(chan error, 1)
	func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- fmt.Errorf("Panicked: %v", r)
			}
		}()
		ch <- wrapped()
	}()
	return <-ch
}
