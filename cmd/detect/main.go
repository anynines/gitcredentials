package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"

	"github.com/avarteqgmbh/gitcredentials/git"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	packit.Detect(git.Detect(logger))
}
