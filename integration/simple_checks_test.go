package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testSimpleChecks(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("run build with and without environment variables", func() {
		var (
			image occam.Image

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("fails without defined environment variables", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().WithVerbose().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Gitcredentials.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).To(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"    Not participating: could not find GIT credentials in environment or in buildpack.yml",
			))
		})

		it("installs with defined environment variables", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Gitcredentials.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{
					"GIT_CREDENTIALS_USERNAME": "testusername",
					"GIT_CREDENTIALS_PASSWORD": "testpassword",
				}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Using environment variables GIT_CREDENTIALS_USERNAME and GIT_CREDENTIALS_PASSWORD",
				"  Initializing GIT credentials cache",
				"    Running command: /usr/bin/git config --replace-all --global credential.helper ' cache --timeout 3600 '",
				"    Command succeeded",
				"",
				"  Configuring git to use HTTPs for authentication",
				"    Running command: /usr/bin/git config --global credential.https://github.com/.username testusername",
				"    Command succeeded",
				"",
				"    Running command: /usr/bin/git config --global url.https://github.com/.insteadOf git@github.com:",
				"    Command succeeded",
				"",
				"  Adding credentials to GIT credentials cache",
				"    Adding credentials succeeded",
			))
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
		})
	})
}
