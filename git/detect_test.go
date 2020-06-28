package git_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/avarteqgmbh/gitcredentials/git"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect            = NewWithT(t).Expect
		logger            scribe.Logger
		detect            packit.DetectFunc
		workingDir        string
		cnbDir            string
		buildPackTomlPath string = "../test/fixtures/some_buildpack.toml"
		buildPackYMLPath  string
	)

	context("when no buildpack.yml is presented", func() {
		it.Before(func() {
			logger = scribe.NewLogger(os.Stdout)
			detect = git.Detect(logger)

			var err error
			workingDir, err = ioutil.TempDir("", "workingDir")
			Expect(err).NotTo(HaveOccurred())

			cnbDir, err = ioutil.TempDir("", "cnb")
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
			Expect(os.RemoveAll(cnbDir)).To(Succeed())
		})

		it("does not participate", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
			})
			Expect(err).To(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{}))
		})

		it("returns a DetectResult", func() {
			someBuildPackTomlFile, err := ioutil.ReadFile(buildPackTomlPath)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
			Expect(err).NotTo(HaveOccurred())

			os.Setenv("GIT_CREDENTIALS_USERNAME", "testuser")
			os.Setenv("GIT_CREDENTIALS_PASSWORD", "testpass")

			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
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
			}))

			os.Unsetenv("GIT_CREDENTIALS_USERNAME")
			os.Unsetenv("GIT_CREDENTIALS_PASSWORD")
		})
	})

	context("when a buildpack.yml is presented", func() {
		it.Before(func() {
			logger = scribe.NewLogger(os.Stdout)
			detect = git.Detect(logger)

			var err error
			workingDir, err = ioutil.TempDir("", "workingDir")
			Expect(err).NotTo(HaveOccurred())

			buildPackYMLPath = filepath.Join(workingDir, "buildpack.yml")
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("does not participate", func() {
			err := ioutil.WriteFile(buildPackYMLPath, []byte("\t"), 0644)
			Expect(err).NotTo(HaveOccurred())

			result, err := detect(packit.DetectContext{WorkingDir: workingDir})
			Expect(err).To(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{}))
		})

		it("returns an empty BuildPackYml struct when the buildpack.yml can be parsed but does not contain a gitcredentials map", func() {
			err := ioutil.WriteFile(buildPackYMLPath, []byte("---"), 0644)
			Expect(err).NotTo(HaveOccurred())

			result, err := detect(packit.DetectContext{WorkingDir: workingDir})
			Expect(err).To(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{}))
		})

		it("returns a DetectResult when the buildpack.yml can be parsed and it contains a gitcredentials map", func() {
			err := ioutil.WriteFile(buildPackYMLPath, []byte(`---
gitcredentials:
  credentials:
    - protocol: https
      host: example.com
      path: /foo.git
      username: username
      password: password
      url: https://example.com
`), 0644)
			Expect(err).NotTo(HaveOccurred())

			result, err := detect(packit.DetectContext{WorkingDir: workingDir})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
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
			}))
		})
	})
}
