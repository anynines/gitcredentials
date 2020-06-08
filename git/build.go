package git

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

// BuildEnvironment represents a build environment for this buildpack
type BuildEnvironment struct {
	BuildPackYML BuildPackYML
	Context      packit.BuildContext
	Logger       scribe.Logger
}

// Build executes the main functionality if this buildpack participates in the
// build plan
func Build(logger scribe.Logger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		envCredentials := GitCredential{}
		gitUserName, userNameExists := os.LookupEnv("GIT_CREDENTIALS_USERNAME")
		gitPassword, passwordExists := os.LookupEnv("GIT_CREDENTIALS_PASSWORD")
		if userNameExists && len(gitUserName) > 0 && passwordExists && len(gitPassword) > 0 {
			logger.Process("Using environment variables GIT_CREDENTIALS_USERNAME and GIT_CREDENTIALS_PASSWORD")

			envCredentials = GitCredential{
				Username: gitUserName,
				Password: gitPassword,
			}

			configuration, err := ReadConfiguration(context.CNBPath)
			if err != nil {
				return packit.BuildResult{}, err
			}

			gitProtocol, protocolExists := os.LookupEnv("GIT_CREDENTIALS_PROTOCOL")
			if protocolExists && len(gitProtocol) > 0 {
				envCredentials.Protocol = gitProtocol
			} else {
				envCredentials.Protocol = configuration.DefaultProcotol
			}

			gitHost, hostExists := os.LookupEnv("GIT_CREDENTIALS_HOST")
			if hostExists && len(gitHost) > 0 {
				envCredentials.Host = gitHost
			} else {
				envCredentials.Host = configuration.DefaultHost
			}

			gitPath, pathExists := os.LookupEnv("GIT_CREDENTIALS_PATH")
			if pathExists && len(gitPath) > 0 {
				envCredentials.Path = gitPath
			} else {
				envCredentials.Path = configuration.DefaultPath
			}

			gitURL, urlExists := os.LookupEnv("GIT_CREDENTIALS_URL")
			if urlExists && len(gitURL) > 0 {
				envCredentials.URL = gitURL
			} else {
				envCredentials.URL = configuration.DefaultURL
			}
		}

		buildPackYML, err := BuildpackYMLParse(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil && !os.IsNotExist(err) {
			return packit.BuildResult{}, err
		}

		if len(envCredentials.Username) > 0 && len(envCredentials.Password) > 0 {
			buildPackYML.Credentials = append(buildPackYML.Credentials, envCredentials)
		}

		if len(buildPackYML.Credentials) == 0 {
			return packit.BuildResult{}, errors.New("No credentials were specified either in environment variables or in the buildpack.yml")
		}

		env := BuildEnvironment{
			BuildPackYML: buildPackYML,
			Context:      context,
			Logger:       logger,
		}

		err = env.Initialize()
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

	for _, credential := range e.BuildPackYML.Credentials {
		credentialURL := credential.Protocol + "://" + credential.Host
		if credential.URL != "" {
			credentialURL = credential.URL
		}

		if credential.Path != "" {
			credentialURL += credential.Path
		} else {
			credentialURL += "/"
		}

		credentialContext := "credential." + credentialURL + ".username"
		err := e.RunGitCommand([]string{
			"git",
			"config",
			"--global",
			credentialContext,
			credential.Username,
		})
		if err != nil {
			return err
		}

		err = e.RunGitCommand([]string{
			"git",
			"config",
			"--global",
			"url." + credentialURL + ".insteadOf",
			"git@" + credential.Host + ":",
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// StoreCredentials runs "git credential approve" to add credentials to the GIT
// credential cache
func (e BuildEnvironment) StoreCredentials() error {
	e.Logger.Process("Adding credentials to GIT credentials cache")

	for _, credential := range e.BuildPackYML.Credentials {
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
			io.WriteString(stdin, "protocol="+credential.Protocol+"\n")
			io.WriteString(stdin, "host="+credential.Host+"\n")
			io.WriteString(stdin, "path="+credential.Path+"\n")
			io.WriteString(stdin, "username="+credential.Username+"\n")
			io.WriteString(stdin, "password="+credential.Password+"\n")
			if credential.URL != "" {
				io.WriteString(stdin, "url="+credential.URL+"\n")
			}
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
	}

	return nil
}
