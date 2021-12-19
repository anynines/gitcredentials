package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"

	"github.com/anynines/gitcredentials/git"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	packit.Build(git.Build(logger))
}
