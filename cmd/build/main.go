package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	"github.com/anynines/gitcredentials/git"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	packit.Build(git.Build(logger))
}
