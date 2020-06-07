package git

import (
	"os"

	yaml "gopkg.in/yaml.v2"
)

// BuildPackYMLCredential represents a git credentials to be stored in the
// Git credentials cache
type BuildPackYMLCredential struct {
	Protocol string `yaml:"protocol"`
	Host     string `yaml:"host"`
	Path     string `yaml:"path"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	URL      string `yaml:"url"`
}

// BuildPackYML represents the buildpack.yml file provided by a user / an app
type BuildPackYML struct {
	Credentials []BuildPackYMLCredential `yaml:"credentials"`
}

// BuildpackYMLParse parses the buildpack.yml file
func BuildpackYMLParse(path string) (BuildPackYML, error) {
	var buildpack struct {
		Gitcredentials BuildPackYML `yaml:"gitcredentials"`
	}

	file, err := os.Open(path)
	if err != nil {
		return BuildPackYML{}, err
	}
	defer file.Close()

	if !os.IsNotExist(err) {
		err = yaml.NewDecoder(file).Decode(&buildpack)
		if err != nil {
			return BuildPackYML{}, err
		}
	}

	return buildpack.Gitcredentials, nil
}
