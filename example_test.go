package shlike

func Example() {
	// Prepare a new config
	cfg := NewConfig()

	// Set some variables
	cfg.Set("FOO", "bar", "baz")
	cfg.Append("BAR", "baz")
	cfg.Append("BAR", "quux")

	// Evaluate a string...
	cfg.Eval("So, $FOO $BAR")

	// One line has been read
	cfg.Length() // returns: 1

	// And it contains variable expansion
	cfg.Line(0) // returns: {"So,", "bar", "baz", "baz", "quux"}

	// load a config file
	cfg.Load("fixtures/example.conf")

	// Read variables set from configuration
	cfg.Get("REDIS_PORT") // returns: {"6379"}

	// Iterate over lines
	for line := range cfg.Iter() {
		_ = line // Process the configuration
	}

	// Save to a file
	cfg.Save("final.conf")
}
