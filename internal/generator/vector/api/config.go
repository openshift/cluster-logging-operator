package api

// Config represents a configuration for vector
type Config struct {
	Global

	// Api is the set of API keys to values
	Api *Api `json:"api,omitempty" yaml:"api,omitempty" toml:"api,omitempty"`

	// Secrets is the set of secret ids to secret configurations
	Secret map[string]*Secret `json:"secret,omitempty" yaml:"secret,omitempty" toml:"secret,omitempty"`

	// Sources is the set of source ids to source configurations
	Sources Sources `json:"sources,omitempty" yaml:"sources,omitempty" toml:"sources,omitempty"`

	// Transforms is the set of transform ids to transform configurations
	Transforms Transforms `json:"transforms,omitempty" yaml:"transforms,omitempty" toml:"transforms,omitempty"`

	Sinks Sinks `json:"sinks,omitempty" yaml:"sinks,omitempty" toml:"sinks,omitempty"`
}

func NewConfig(init func(*Config)) *Config {
	c := &Config{
		Secret:     make(map[string]*Secret),
		Sources:    make(Sources),
		Transforms: make(Transforms),
		Sinks:      make(Sinks),
	}
	if init != nil {
		init(c)
	}
	return c
}

func (c *Config) AddSinks(sinks Sinks) {
	for id, s := range sinks {
		c.Sinks[id] = s
	}
}

func (c *Config) AddSources(sources Sources) {
	for id, s := range sources {
		c.Sources[id] = s
	}
}

func (c *Config) AddTransforms(transforms Transforms) {
	for id, t := range transforms {
		c.Transforms[id] = t
	}
}
