package shlike

type DotCommand func(*Config, ...string) error

func DotInclude(c *Config, paths ...string) error {
	for _, path := range paths {
		if err := c.Load(path); err != nil {
			return err
		}
	}
	return nil
}

var DotCommands = map[string]DotCommand{"include": DotInclude}
