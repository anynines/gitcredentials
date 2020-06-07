package git

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

// Detect determines whether this buildpack should participate
func Detect(logger scribe.Logger) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		_, err := BuildpackYMLParse(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil && os.IsNotExist(err) {
			return packit.DetectResult{}, err
		}

		if err != nil {
			return packit.DetectResult{}, errors.New("GIT credentials cannot be found in buildpack.yml")
		}

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

		// gitUserName, userNameExists := os.LookupEnv("GIT_USERNAME")
		// gitToken, tokenExists := os.LookupEnv("GIT_TOKEN")
		// if userNameExists && len(gitUserName) > 0 && tokenExists && len(gitToken) > 0 {
		// 	logger.Process("Found environment variables GIT_USERNAME and GIT_TOKEN")

		// 	return packit.DetectResult{
		// 		Plan: packit.BuildPlan{
		// 			Provides: []packit.BuildPlanProvision{
		// 				{Name: "gitcredentials"},
		// 			},
		// 			Requires: []packit.BuildPlanRequirement{
		// 				{
		// 					Name: "gitcredentials",
		// 				},
		// 			},
		// 		},
		// 	}, nil
		// }

		// return packit.DetectResult{}, errors.New("Environment variables GIT_USERNAME and GIT_TOKEN not found")
	}
}
