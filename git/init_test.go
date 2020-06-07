package git_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitGitCredentials(t *testing.T) {
	suite := spec.New("gitcredentials", spec.Report(report.Terminal{}))
	suite("Configuration", testConfiguration)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Detect", testDetect)
	suite.Run(t)
}
