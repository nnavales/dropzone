package config

// Default returns the initial config used when no config file exists.
func Default() Config {
	cfg := Config{
		Settings: Settings{
			StableForSeconds: 2,
			OnConflict:       "rename",
		},
	}
	return cfg
}

func (c *Config) withDefaults() {
	if c.Settings.StableForSeconds == 0 {
		c.Settings.StableForSeconds = 2
	}
	if c.Settings.OnConflict == "" {
		c.Settings.OnConflict = "rename"
	}
	for i := range c.Zones {
		for j := range c.Zones[i].Rules {
			if c.Zones[i].Rules[j].OnConflict == "" {
				c.Zones[i].Rules[j].OnConflict = c.Settings.OnConflict
			}
		}
	}
}
