package shlike

import "os"
import "testing"
import . "github.com/smartystreets/goconvey/convey"

func TestConfig(t *testing.T) {
	Convey("Configuration", t, func() {
		var cfg = NewConfig()

		Convey("Variables", func() {
			Convey("Setting", func() {
				cfg.Set("FOO")
				cfg.Set("BAR", "baz")
				cfg.Set("QUUX", "xyzzy", "barney")

				So(cfg.Vars, ShouldResemble, map[string][]string{
					"FOO":  []string{},
					"BAR":  []string{"baz"},
					"QUUX": []string{"xyzzy", "barney"},
				})
			})

			Convey("Getting", func() {
				cfg.Set("FOO", "bar", "baz")
				So(cfg.Get("FOO"), ShouldResemble, []string{"bar", "baz"})
				So(cfg.Get("BAR"), ShouldBeNil)
			})

			Convey("Unsetting", func() {
				cfg.Set("FOO", "bar")
				cfg.Unset("FOO")
				So(cfg.Get("FOO"), ShouldBeNil)
			})

			Convey("Appending", func() {
				cfg.Append("FOO", "bar", "baz")
				cfg.Append("FOO", "quux")
				cfg.Append("FOO", "xyzzy")
				So(cfg.Get("FOO"), ShouldResemble, []string{"bar", "baz", "quux", "xyzzy"})
			})
		})

		Convey("Line access", func() {
			cfg.Eval("foo\nbar\nbaz")
			So(cfg.Line(0), ShouldResemble, []string{"foo"})
			So(cfg.Line(1), ShouldResemble, []string{"bar"})
			So(cfg.Line(2), ShouldResemble, []string{"baz"})
			So(cfg.Line(3), ShouldBeNil)
		})

		Convey("Load", func() {
			So(cfg.Load("fixtures/example.conf"), ShouldBeNil)
			So(cfg.Get("PGPASSWORD"), ShouldResemble, []string{"dupa.8"})

			Convey("Returns error when appropriate", func() {
				So(cfg.Load("fixtures/nonexistent.conf"), ShouldNotBeNil)
			})
		})

		Convey("Serialize", func() {
			Convey("Works at all", func() {
				cfg.Eval("FOO = 1\nTony Halik")
				So(cfg.Serialize(), ShouldEqual, "FOO = 1\nTony Halik")
			})

			Convey("Can be evaluated back", func() {
				cfg.Load("fixtures/example.conf")
				reloaded := NewConfig()
				So(reloaded.Eval(cfg.Serialize()), ShouldBeNil)
				So(reloaded.Vars, ShouldResemble, cfg.Vars)
				So(reloaded.Lines, ShouldResemble, cfg.Lines)
			})
		})

		Convey("Save", func() {
			os.MkdirAll("tmp", 0700)
			cfg.Load("fixtures/example.conf")
			So(cfg.Save("tmp/saved.conf"), ShouldBeNil)
			reloaded := NewConfig()
			So(reloaded.Load("tmp/saved.conf"), ShouldBeNil)
			So(reloaded.Vars, ShouldResemble, cfg.Vars)
			So(reloaded.Lines, ShouldResemble, cfg.Lines)
		})
	})
}
