package git

import (
	"os"

	yaml "gopkg.in/yaml.v2"
)

// BuildPackYML represents the buildpack.yml file provided by a user / an app
type BuildPackYML struct {
	Contexts string `yaml:"contexts"`
}

// BuildpackYMLParser represents the buildpack.yml parser
type BuildpackYMLParser struct{}

// NewBuildpackYMLParser creates and returns a new buildpack.yml parser
func NewBuildpackYMLParser() BuildpackYMLParser {
	return BuildpackYMLParser{}
}

// BuildpackYMLParse parses the buildpack.yml file
func BuildpackYMLParse(path string) (BuildPackYML, error) {
	var buildpack struct {
		Rvm BuildPackYML `yaml:"git"`
	}

	file, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return BuildPackYML{}, err
	}
	defer file.Close()

	if !os.IsNotExist(err) {
		err = yaml.NewDecoder(file).Decode(&buildpack)
		if err != nil {
			return BuildPackYML{}, err
		}
	}

	return buildpack.Rvm, nil
}

// ParseVersion parses the buildpack.yml file and returns a a ruby version, if
// one was specified
func (p BuildpackYMLParser) ParseVersion(path string) (string, error) {
	config, err := BuildpackYMLParse(path)
	if err != nil {
		return "", err
	}

	return config.Contexts, nil
}
