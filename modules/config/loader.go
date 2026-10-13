package config

// Loader loads application Config from a source.
type Loader interface {
	Load() (*Config, error)
}
