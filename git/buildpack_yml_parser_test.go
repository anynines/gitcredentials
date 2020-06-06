package git_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildpackYMLParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser git.BuildpackYMLParser
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "buildpack.yml")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(`---
rvm:
  ruby_version: 2.6.1
`)
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()

		parser = git.NewBuildpackYMLParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Parse", func() {
		it.Before(func() {
			err := ioutil.WriteFile(path, []byte(`---
rvm:
  rvm_version: 1.29.9
  ruby_version: 2.6.1
  node_version: 10.*
  require_node: false
`), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		it("parses a buildpack.yml file", func() {
			configData, err := git.BuildpackYMLParse(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(configData.RvmVersion).To(Equal("1.29.9"))
			Expect(configData.RubyVersion).To(Equal("2.6.1"))
			Expect(configData.NodeVersion).To(Equal("10.*"))
			Expect(configData.RequireNode).To(BeFalse())
		})
	})

	context("ParseVersion", func() {
		it("parses the node version from a buildpack.yml file", func() {
			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("2.6.1"))
		})

		context("when the buildpack.yml file does not exist", func() {
			it.Before(func() {
				Expect(os.Remove(path)).To(Succeed())
			})

			it("returns an empty version", func() {
				version, err := parser.ParseVersion(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(version).To(BeEmpty())
			})
		})

		context("failure cases", func() {
			context("when the buildpack.yml file cannot be read", func() {
				it.Before(func() {
					Expect(os.Chmod(path, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(path, 0644)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when the contents of the buildpack.yml file are malformed", func() {
				it.Before(func() {
					err := ioutil.WriteFile(path, []byte("%%%"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError(ContainSubstring("could not find expected directive name")))
				})
			})
		})
	})
}
