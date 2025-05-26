package configs

type Core struct {
	TickRate    int    `yaml:"tick_rate"`
	DefaultArea string `yaml:"default_area"`
	DefaultRoom string `yaml:"default_room"`
}

func (c *Core) Check() {
	if c.TickRate == 0 {
		c.TickRate = 100
	}
}
