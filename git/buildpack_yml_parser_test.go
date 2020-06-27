package git_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/avarteqgmbh/gitcredentials/git"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildpackYMLParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		path   string
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "buildpack.yml")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(`---
gitcredentials:
  credentials:
    - protocol: https
      host: example.com
      path: /foo.git
      username: username
      password: password
      url: https://example.com
`)
		Expect(err).NotTo(HaveOccurred())
		path = file.Name()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Parse", func() {
		it("parses a buildpack.yml file", func() {
			gitcredentials, err := git.BuildpackYMLParse(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(gitcredentials.Credentials[0].Protocol).To(Equal("https"))
			Expect(gitcredentials.Credentials[0].Host).To(Equal("example.com"))
			Expect(gitcredentials.Credentials[0].Path).To(Equal("/foo.git"))
			Expect(gitcredentials.Credentials[0].Username).To(Equal("username"))
			Expect(gitcredentials.Credentials[0].Password).To(Equal("password"))
			Expect(gitcredentials.Credentials[0].URL).To(Equal("https://example.com"))
		})
	})

	context("Parsing errors", func() {
		context("when the buildpack.yml file does not exist", func() {
			it.Before(func() {
				Expect(os.Remove(path)).To(Succeed())
			})

			it("returns an empty YAML structure", func() {
				gitcredentials, err := git.BuildpackYMLParse(path)
				Expect(err).To(HaveOccurred())
				Expect(gitcredentials).To(Equal(git.BuildPackYML{Credentials: nil}))
			})
		})

		context("when the buildpack.yml file does not contain a gitcredentials map", func() {
			it.Before(func() {
				err := ioutil.WriteFile(path, []byte("---"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an empty BuildPackYML structure", func() {
				gitcredentials, err := git.BuildpackYMLParse(path)
				Expect(err).To(Not(HaveOccurred()))
				Expect(gitcredentials).To(Equal(git.BuildPackYML{}))
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
					_, err := git.BuildpackYMLParse(path)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when the contents of the buildpack.yml file are malformed", func() {
				it.Before(func() {
					err := ioutil.WriteFile(path, []byte("%%%"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := git.BuildpackYMLParse(path)
					Expect(err).To(MatchError(ContainSubstring("could not find expected directive name")))
				})
			})
		})
	})
}
