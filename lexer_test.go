package shlike

import "io/ioutil"
import "os"
import "testing"

import . "github.com/smartystreets/goconvey/convey"

func stderrFor(fn func()) string {
	defer func(stderr *os.File) { os.Stderr = stderr }(os.Stderr)
	os.Stderr, _ = ioutil.TempFile("", "shlike.test.")
	fn()
	os.Stderr.Sync()
	os.Stderr.Seek(0, 0)
	stderrb, _ := ioutil.ReadAll(os.Stderr)
	return string(stderrb)
}

func TestLexer(t *testing.T) {
	Convey("Lexical Analysis", t, func() {
		var c = NewConfig()

		Convey("Word and line splitting and escaping", func() {
			Convey("Smoke test", func() {
				c.Eval(`
foo bar
baz
quux
tony halik
`)
				So(c.lines, ShouldResemble, [][]string{
					{"foo", "bar"},
					{"baz"},
					{"quux"},
					{"tony", "halik"},
				})
			})

			Convey("Escaping", func() {
				c.Eval(`
Tony\ Halik 'Tony Halik' "Tony Halik" \
\T'o'"n"y' H'alik
'$tota#lly' \$what\$0\#\"ever\' '"' "'"
'foo
bar' "baz
quux" "xy\
zzy" 'what\
ever'
`)
				c.Eval("cr\r\nlf\\\r\n\"cr\\\r\nlf\"")
				So(c.lines, ShouldResemble, [][]string{
					{"Tony Halik", "Tony Halik", "Tony Halik", "Tony Halik"},
					{"$tota#lly", "$what$0#\"ever'", "\"", "'"},
					{"foo\nbar", "baz\nquux", "xyzzy", "what\\\never"},
					{"cr"}, {"lf", "crlf"},
				})
			})

			Convey("Empty strings", func() {
				c.Eval("Tony Hal\"\"ik '' Tony\\ Ha''lik \"\" ")
				So(c.lines, ShouldResemble, [][]string{{"Tony", "Halik", "", "Tony Halik", ""}})
			})

			Convey("Comments", func() {
				c.Eval("Tony Halik # tu byłem")
				So(c.lines, ShouldResemble, [][]string{{"Tony", "Halik"}})
			})
		})

		Convey("Variable setting", func() {
			c.Set("FOO", "Foo")
			c.Set("BAR", "Bar")
			c.Set("BAZ", "Baz")
			c.Eval(`
FOO = Tony Halik
BAR ?= Tony Halik
BAZ += Quux
QUUX = Quux
`)
			So(c.Get("FOO"), ShouldResemble, []string{"Tony", "Halik"})
			So(c.Get("BAR"), ShouldResemble, []string{"Bar"})
			So(c.Get("BAZ"), ShouldResemble, []string{"Baz", "Quux"})
			So(c.Get("QUUX"), ShouldResemble, []string{"Quux"})
		})

		Convey("Variable expansion", func() {
			c.Set("FOO", "Tony", "Halik")
			c.Eval(`$FOO "$FOO" '$FOO' ${FOO} "${FOO}" '${FOO}' tu${FOO}byłem "tu${FOO}byłem" tam$FOO-też "tam$FOO-też"`)
			So(c.lines, ShouldResemble, [][]string{{
				"Tony", "Halik", // bare/unquoted
				"Tony Halik",    // bare/doublequoted
				"$FOO",          // bare/singlequoted
				"Tony", "Halik", // braced/unquoted
				"Tony Halik",                   // braced/doublequoted
				"${FOO}",                       // braced/singlequoted
				"tu", "Tony", "Halik", "byłem", // braced, surronded, unquoted
				"tuTony Halikbyłem",            // braced, surrounded, doublequoted
				"tam", "Tony", "Halik", "-też", // unbraced, surrounded, unquoted
				"tamTony Halik-też", // unbraced, surrounded, doublequoted
			}})
		})

		Convey("Invalid input", func() {
			So(c.Eval(`'foo`), ShouldNotBeNil)
			So(c.Eval(`"foo`), ShouldNotBeNil)
			So(c.Eval("fo\ro"), ShouldNotBeNil)
		})

		Convey("Undefined variable warnings", func() {
			So(stderrFor(func() { c.Eval("$undef") }), ShouldEndWith, "WARNING: Undefined variable \"undef\"\n")
		})

		Convey("Debug function (for full coverage)", func() {
			l := c.lexer("", "")
			So(stderrFor(func() { l.debug("foo") }), ShouldContainSubstring, "foo")
		})
	})
}
