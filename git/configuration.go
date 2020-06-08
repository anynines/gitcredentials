package git

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Configuration represents this buildpack's configuration read from a table
// named "configuration"
type Configuration struct {
	DefaultTimeout  string `toml:"default_timeout"`
	DefaultProcotol string `toml:"default_protocol"`
	DefaultHost     string `toml:"default_host"`
	DefaultPath     string `toml:"default_path"`
	DefaultURL      string `toml:"default_url"`
}

// MetaData represents this buildpack's metadata
type MetaData struct {
	Metadata struct {
		Configuration Configuration `toml:"configuration"`
	} `toml:"metadata"`
}

// ReadConfiguration returns the configuration for this buildpack
func ReadConfiguration(cnbPath string) (Configuration, error) {
	file, err := os.Open(filepath.Join(cnbPath, "buildpack.toml"))
	if err != nil {
		return Configuration{}, err
	}

	var meta MetaData
	_, err = toml.DecodeReader(file, &meta)
	if err != nil {
		return Configuration{}, err
	}

	return meta.Metadata.Configuration, nil
}
