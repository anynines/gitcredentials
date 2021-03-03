package git_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/avarteqgmbh/gitcredentials/git"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testConfiguration(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cnbDir            string
		buildPackTomlPath string = "../test/fixtures/some_buildpack.toml"
	)

	context("when the buildpack configuration is requested", func() {
		it.Before(func() {
			var err error
			cnbDir, err = ioutil.TempDir("", "cnb")
			Expect(err).NotTo(HaveOccurred())
		})

		it("cannot find the buildpack.toml file and returns an error", func() {
			_, err := git.ReadConfiguration(cnbDir)
			Expect(err).To(HaveOccurred())
		})

		it("reads an invalid buildpack.toml file and returns an error", func() {
			err := ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte("[[buildpack]"), 0644)
			Expect(err).NotTo(HaveOccurred())

			_, err = git.ReadConfiguration(cnbDir)
			Expect(err).To(HaveOccurred())
		})

		it("reads a valid buildpack.toml file without the [metadata.configuration] table and returns an empty configuration", func() {
			err := ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte("[buildpack]"), 0644)
			Expect(err).NotTo(HaveOccurred())

			configuration, _ := git.ReadConfiguration(cnbDir)
			Expect(configuration).To(Equal(git.Configuration{}))
		})

		it("reads the buildpack.toml file and returns the configuration", func() {
			someBuildPackTomlFile, err := ioutil.ReadFile(buildPackTomlPath)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
			Expect(err).NotTo(HaveOccurred())

			configuration, err := git.ReadConfiguration(cnbDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(configuration).To(Equal(git.Configuration{
				DefaultTimeout:  "3600",
				DefaultProcotol: "https",
				DefaultHost:     "github.com",
				DefaultPath:     "/",
				DefaultURL:      "",
			}))
		})

		it.After(func() {
			Expect(os.RemoveAll(cnbDir)).To(Succeed())
		})
	})
}
