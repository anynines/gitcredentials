package git

import (
	"errors"
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

// Detect determines whether this buildpack should participate
func Detect(logger scribe.Logger, buildpackYMLParser BuildpackYMLParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		gitUserName, userNameExists := os.LookupEnv("GIT_USERNAME")
		gitToken, tokenExists := os.LookupEnv("GIT_TOKEN")
		if userNameExists && len(gitUserName) > 0 && tokenExists && len(gitToken) > 0 {
			logger.Process("Found environment variables GIT_USERNAME and GIT_TOKEN")

			return packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "gitcredentials"},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "gitcredentials",
						},
					},
				},
			}, nil
		}

		return packit.DetectResult{}, errors.New("Environment variables GIT_USERNAME and GIT_TOKEN not found")
	}
}
