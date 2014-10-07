package shlike

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

				So(cfg.variables, ShouldResemble, map[string][]string{
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

		Convey("Can be loaded from a file", func() {
			So(cfg.Load("fixtures/example.conf"), ShouldBeNil)

			Convey("Will fail on loading errors", func() {
				So(cfg.Load("fixtures/nonexistent.conf"), ShouldNotBeNil)
			})
		})

		Convey("Serialization", func() {
			Convey("Works at all", func() {
				cfg.Eval(`
FOO = 1
Tony Halik
`)
				So(cfg.String(), ShouldEqual, "FOO = 1\nTony Halik")
			})

			Convey("Serialized config can be loaded back", func() {
				cfg.Load("fixtures/example.conf")
				reloaded := NewConfig()
				reloaded.Eval(cfg.String())
				So(reloaded, ShouldResemble, cfg)
			})
		})
	})
}
