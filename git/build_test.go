package git_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/avarteqgmbh/gitcredentials/git"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect                   = NewWithT(t).Expect
		logger                   scribe.Logger
		build                    packit.BuildFunc
		workingDir               string
		cnbDir                   string
		buildPackTomlPath        string = "../test/fixtures/some_buildpack.toml"
		invalidBuildPackTomlPath string = "../test/fixtures/invalid_buildpack.toml"
	)

	it.Before(func() {
		var err error
		logger = scribe.NewLogger(os.Stdout)
		build = git.Build(logger)

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("returns a BuildResult", func() {
		someBuildPackTomlFile, err := ioutil.ReadFile(buildPackTomlPath)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("GIT_CREDENTIALS_USERNAME", "testuser")
		os.Setenv("GIT_CREDENTIALS_PASSWORD", "testpass")

		result, err := build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "rvm-bundler",
						Metadata: map[string]interface{}{
							"version": "0.0.x",
						},
					},
				},
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "rvm-bundler",
						Metadata: map[string]interface{}{
							"version": "0.0.x",
						},
					},
				},
			},
			Layers: []packit.Layer{
				{
					Path:      "gitcredentials",
					Name:      "gitcredentials",
					Build:     false,
					Launch:    false,
					Cache:     false,
					SharedEnv: packit.Environment{},
					BuildEnv:  packit.Environment{},
					LaunchEnv: packit.Environment{},
					Metadata:  nil,
				},
			},
			Launch: packit.LaunchMetadata{Processes: nil, Slices: nil, Labels: nil},
		}))

		os.Unsetenv("GIT_CREDENTIALS_USERNAME")
		os.Unsetenv("GIT_CREDENTIALS_PASSWORD")
		unsetGitCredentials()
	})

	it("environment variables are not set", func() {
		someBuildPackTomlFile, err := ioutil.ReadFile(buildPackTomlPath)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
		Expect(err).NotTo(HaveOccurred())

		_, err = build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "rvm-bundler",
						Metadata: map[string]interface{}{
							"version": "0.0.x",
						},
					},
				},
			},
		})
		Expect(err).To(MatchError("No credentials were specified either in environment variables or in the buildpack.yml"))
	})

	it("all environment variables are set", func() {
		someBuildPackTomlFile, err := ioutil.ReadFile(buildPackTomlPath)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("GIT_CREDENTIALS_USERNAME", "testuser")
		os.Setenv("GIT_CREDENTIALS_PASSWORD", "testpass")
		os.Setenv("GIT_CREDENTIALS_PROTOCOL", "testprotocol")
		os.Setenv("GIT_CREDENTIALS_HOST", "testhost.com")
		os.Setenv("GIT_CREDENTIALS_PATH", "/testpath/")
		os.Setenv("GIT_CREDENTIALS_URL", "https://newexample.com")

		result, err := build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "rvm-bundler",
						Metadata: map[string]interface{}{
							"version": "0.0.x",
						},
					},
				},
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "rvm-bundler",
						Metadata: map[string]interface{}{
							"version": "0.0.x",
						},
					},
				},
			},
			Layers: []packit.Layer{
				{
					Path:      "gitcredentials",
					Name:      "gitcredentials",
					Build:     false,
					Launch:    false,
					Cache:     false,
					SharedEnv: packit.Environment{},
					BuildEnv:  packit.Environment{},
					LaunchEnv: packit.Environment{},
					Metadata:  nil,
				},
			},
			Launch: packit.LaunchMetadata{Processes: nil, Slices: nil, Labels: nil},
		}))

		os.Unsetenv("GIT_CREDENTIALS_USERNAME")
		os.Unsetenv("GIT_CREDENTIALS_PASSWORD")
		os.Unsetenv("GIT_CREDENTIALS_PROTOCOL")
		os.Unsetenv("GIT_CREDENTIALS_HOST")
		os.Unsetenv("GIT_CREDENTIALS_PATH")
		os.Unsetenv("GIT_CREDENTIALS_URL")
		unsetGitCredentials()
	})

	it("git binary is not installed", func() {
		someBuildPackTomlFile, err := ioutil.ReadFile(buildPackTomlPath)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("GIT_CREDENTIALS_USERNAME", "testuser")
		os.Setenv("GIT_CREDENTIALS_PASSWORD", "testpass")
		path := os.Getenv("PATH")
		os.Unsetenv("PATH")

		_, err = build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "rvm-bundler",
						Metadata: map[string]interface{}{
							"version": "0.0.x",
						},
					},
				},
			},
		})
		Expect(err).To(MatchError("exec: \"git\": executable file not found in $PATH"))

		os.Setenv("PATH", path)
		os.Unsetenv("GIT_CREDENTIALS_USERNAME")
		os.Unsetenv("GIT_CREDENTIALS_PASSWORD")
	})

	it("invalid buildpack.toml", func() {
		someBuildPackTomlFile, err := ioutil.ReadFile(invalidBuildPackTomlPath)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), someBuildPackTomlFile, 0644)
		Expect(err).NotTo(HaveOccurred())

		_, err = build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "rvm-bundler",
						Metadata: map[string]interface{}{
							"version": "0.0.x",
						},
					},
				},
			},
		})
		Expect(err).To(MatchError("Near line 0 (last key parsed ''): unexpected end of table name (table names cannot be empty)"))
	})
}

func unsetGitCredentials() {
	cmds := [][]string{
		{"git", "config", "--global", "--unset", "credential.helper"},
		{"git", "config", "--global", "--unset", "credential.https://github.com/.username"},
		{"git", "config", "--global", "--unset", "credential.https://newexample.com/testpath/.username"},
		{"git", "config", "--global", "--unset", "url.https://github.com/.insteadof"},
		{"git", "config", "--global", "--unset", "url.https://newexample.com/testpath/.insteadof"},
	}

	gitbin, err := exec.LookPath("git")
	if err == nil {
		for _, i := range cmds {
			cmd := exec.Command(gitbin)
			cmd.Args = i
			_ = cmd.Run()
		}
	}
}
