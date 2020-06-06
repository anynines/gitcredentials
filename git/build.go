package git

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

// BuildEnvironment represents a build environment for this buildpack
type BuildEnvironment struct {
	BuildpackYMLParser BuildpackYMLParser
	Context            packit.BuildContext
	Logger             scribe.Logger
}

// Build executes the main functionality if this buildpack participates in the
// build plan
func Build(logger scribe.Logger, buildpackYMLParser BuildpackYMLParser) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		env := BuildEnvironment{
			BuildpackYMLParser: buildpackYMLParser,
			Context:            context,
			Logger:             logger,
		}

		err := env.Initialize()
		if err != nil {
			return packit.BuildResult{}, err
		}

		err = env.Configure()
		if err != nil {
			return packit.BuildResult{}, err
		}

		err = env.StoreCredentials()
		if err != nil {
			return packit.BuildResult{}, err
		}

		gitCredentialsLayer, err := context.Layers.Get("gitcredentials", packit.LaunchLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		gitCredentialsLayer.Build = false
		gitCredentialsLayer.Cache = false
		gitCredentialsLayer.Launch = false

		return packit.BuildResult{
			Plan: context.Plan,
			Layers: []packit.Layer{
				gitCredentialsLayer,
			},
		}, nil
	}
}

// RunGitCommand executes a GIT command with given arguments
func (e BuildEnvironment) RunGitCommand(args []string) error {
	cmd := exec.Command("git")
	cmd.Args = args

	e.Logger.Subprocess("Running command: " + cmd.String())

	var stdOutBytes bytes.Buffer
	cmd.Stdout = &stdOutBytes

	var stdErrBytes bytes.Buffer
	cmd.Stderr = &stdErrBytes

	err := cmd.Run()
	if err != nil {
		e.Logger.Subprocess("Command failed")
		if stdErrBytes.Len() > 0 {
			e.Logger.Subprocess("Command stderr: %s", stdErrBytes.String())
		}
		e.Logger.Subprocess("Error status code: %s", err.Error())
		e.Logger.Break()
		return err
	}

	e.Logger.Subprocess("Command succeeded")
	if stdOutBytes.Len() > 0 {
		e.Logger.Subprocess("Command output: %s", stdOutBytes.String())
	}
	e.Logger.Break()

	return nil
}

// Initialize initalizes the GIT credential cache which stores credentials in
// memory exclusively
func (e BuildEnvironment) Initialize() error {
	e.Logger.Process("Initializing GIT credentials cache")
	return e.RunGitCommand([]string{
		"git",
		"config",
		"--global",
		"credential.helper",
		"cache",
	})
}

// Configure creates configuration to direct GIT to use the HTTPs protocol
// rather than the SSH protocol
func (e BuildEnvironment) Configure() error {
	e.Logger.Process("Configuring git to use HTTPs for authentication")

	// git config credential.https://example.com.username myusername
	err := e.RunGitCommand([]string{
		"git",
		"config",
		"--global",
		"credential.https://github.com/.username",
		os.Getenv("GIT_USERNAME"),
	})
	if err != nil {
		return err
	}

	return e.RunGitCommand([]string{
		"git",
		"config",
		"--global",
		"url.https://github.com/.insteadOf",
		"git@github.com:",
	})
}

// StoreCredentials runs "git credential approve" to add credentials to the GIT
// credential cache
func (e BuildEnvironment) StoreCredentials() error {
	e.Logger.Process("Adding credentials to GIT credentials cache")
	cmd := exec.Command("git")
	cmd.Args = []string{
		"git",
		"credential",
		"approve",
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "protocol=https\n")
		io.WriteString(stdin, "host=github.com\n")
		io.WriteString(stdin, "path=/\n")
		io.WriteString(stdin, "username="+os.Getenv("GIT_USERNAME")+"\n")
		io.WriteString(stdin, "password="+os.Getenv("GIT_TOKEN")+"\n")
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	if err != nil {
		e.Logger.Subprocess("Adding credentials failed")
		e.Logger.Subprocess("Error status code: %s", err.Error())

		var stderrBytes []byte
		stderrBytes, err = ioutil.ReadAll(stderr)
		if err == nil && len(stderrBytes) > 0 {
			e.Logger.Subprocess("Command stderr: %s", string(stderrBytes))
		}
		e.Logger.Break()
		return err
	}

	e.Logger.Subprocess("Adding credentials succeeded")

	var stdoutBytes []byte
	stdoutBytes, err = ioutil.ReadAll(stdout)
	if err == nil && len(stdoutBytes) > 0 {
		e.Logger.Subprocess("Command output: %s", string(stdoutBytes))
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	e.Logger.Break()

	return nil
}
