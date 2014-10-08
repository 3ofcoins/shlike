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

				So(c.Lines, ShouldResemble, [][]string{
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
				So(c.Lines, ShouldResemble, [][]string{
					{"Tony Halik", "Tony Halik", "Tony Halik", "Tony Halik"},
					{"$tota#lly", "$what$0#\"ever'", "\"", "'"},
					{"foo\nbar", "baz\nquux", "xyzzy", "what\\\never"},
					{"cr"}, {"lfcrlf"},
				})
			})

			Convey("Empty strings", func() {
				c.Eval("Tony Hal\"\"ik '' Tony\\ Ha''lik \"\" ")
				So(c.Lines, ShouldResemble, [][]string{{"Tony", "Halik", "", "Tony Halik", ""}})
			})

			Convey("Comments", func() {
				c.Eval("Tony Halik # tu byłem")
				So(c.Lines, ShouldResemble, [][]string{{"Tony", "Halik"}})
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
			c.Eval(`$FOO "$FOO" '$FOO' ${FOO} "${FOO}" '${FOO}' tu${FOO}byłem "tu${FOO}byłem" tam$FOO-też "tam$FOO-też" "${FOO|+}"`)
			So(c.Lines, ShouldResemble, [][]string{{
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
				"Tony+Halik",        // braced, doublequoted, custom join<
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

		Convey("Dot-include", func() {
			Convey("Existing file", func() {
				c.Set("PGPASSWORD", "dupa.7")
				c.Set("REDIS_PORT", "6380")
				So(c.Eval(`. fixtures/example.conf`), ShouldBeNil)
				So(c.Get("PGPASSWORD"), ShouldResemble, []string{"dupa.8"})
				So(c.Get("REDIS_PORT"), ShouldResemble, []string{"6380"})
			})

			Convey("Relative path", func() {
				So(c.Eval(`. fixtures/outer.conf`), ShouldBeNil)
				So(c.Get("PGPASSWORD"), ShouldResemble, []string{"dupa.8"})
			})

			Convey("Error", func() {
				So(c.Eval(`. fixtures/nonexistent.conf`), ShouldNotBeNil)
			})

			Convey("Invalid", func() {
				So(c.Eval(`. foo bar`), ShouldNotBeNil)
			})
		})

		Convey("Full coverage", func() {
			l := newLexer(c, "", "")
			So(stderrFor(func() { l.debug("foo") }), ShouldContainSubstring, "foo")
			So(func() { l.op = opKind(-1); l.endLine() }, ShouldPanic)
		})

		Convey("README.md example", func() {
			So(c.Load("fixtures/readme.conf"), ShouldBeNil)
			So(c.Vars, ShouldResemble, map[string][]string{
				"META":     []string{"foo", "bar", "baz", "quux"},
				"NUMBERS":  []string{"4", "8", "15", "16", "23", "42"},
				"SENTENCE": []string{"Lorem ipsum dolor sit amet"}})
			So(c.Lines, ShouldResemble, [][]string{
				[]string{"one", "two", "three?"},
				[]string{"Meta is:", "foo", "bar", "baz", "quux"},
				[]string{"Numbers are: \"4, 8, 15, 16, 23, 42\""},
				[]string{"Words not separated by whitespace are joined together."},
				[]string{"Not", "expanded:", "$META", "${META}"},
				[]string{"single quotes\\ retain\\\nbackslashes and $character"},
				[]string{"double quotes interprete them"},
				[]string{"Backslash discards line breaks"}})
		})
	})
}
