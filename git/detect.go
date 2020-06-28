package git

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

// Detect determines whether this buildpack should participate
func Detect(logger scribe.Logger) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		detectResult := packit.DetectResult{
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
		}

		gitUserName, userNameExists := os.LookupEnv("GIT_CREDENTIALS_USERNAME")
		gitPassword, passwordExists := os.LookupEnv("GIT_CREDENTIALS_PASSWORD")
		if userNameExists && len(gitUserName) > 0 && passwordExists && len(gitPassword) > 0 {
			logger.Process("Using environment variables GIT_CREDENTIALS_USERNAME and GIT_CREDENTIALS_PASSWORD")

			configuration, err := ReadConfiguration(context.CNBPath)
			if err != nil {
				return packit.DetectResult{}, err
			}

			gitProtocol, protocolExists := os.LookupEnv("GIT_CREDENTIALS_PROTOCOL")
			protocolDefined := (protocolExists && len(gitProtocol) > 0) || len(configuration.DefaultProcotol) > 0

			gitHost, hostExists := os.LookupEnv("GIT_CREDENTIALS_HOST")
			hostDefined := (hostExists && len(gitHost) > 0) || len(configuration.DefaultHost) > 0

			gitPath, pathExists := os.LookupEnv("GIT_CREDENTIALS_PATH")
			pathDefined := (pathExists && len(gitPath) > 0) || len(configuration.DefaultPath) > 0

			if protocolDefined && hostDefined && pathDefined {
				return detectResult, nil
			}
		}

		BuildpackYML, err := BuildpackYMLParse(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err == nil && len(BuildpackYML.Credentials) > 0 {
			return detectResult, nil
		}

		logger.Subprocess("Not participating: could not find GIT credentials in environment or in buildpack.yml")
		logger.Break()
		return packit.DetectResult{}, packit.Fail
	}
}
