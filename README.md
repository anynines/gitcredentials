# GIT credentials Cloud native Buildpack

GIT credentials is a Cloud Native Buildpack that allows an app developer to supply credentials for GIT repositories which require authentication.

## How it works

A user of this buildpack can supply a file called `buildpack.yml` in the root directory of the application. The buildpack.yml may contain a map structured as follows:

```yaml
gitcredentials:
  credentials:
    - protocol: https
      host: example.com
      path: /foo.git
      username: username
      password: password
      url: https://example.com

    - protocol: https
      host: example.org
      path: /
      username: other_username
      password: other_password
```

Please read [git-credential](https://git-scm.com/docs/git-credential) to learn more about the semantics of the fields specified in "credentials". Currently, the only supported protocol is HTTPs. Support for SSH is planned.

## What it does

1. If (and only if) the `gitcredentials.credentials` array is found in `buildpack.yml` will [GIT's credential cache](https://git-scm.com/docs/gitcredentials) by initialized by this buildpack. The GIT credential cache stores credentials [exclusively in memory](https://git-scm.com/book/en/v2/Git-Tools-Credential-Storage) (and forgets them after a default timeout of 15 minutes).
1. In addition to that, it sets a [credential context](https://git-scm.com/docs/gitcredentials#_credential_contexts) so that git knows which credentials to use for which host.
1. Lastly, it sets [`url.<base>.insteadOf`](https://git-scm.com/docs/git-config#Documentation/git-config.txt-urlltbasegtinsteadOf) to direct GIT to authenticate using HTTPs instead of SSH. Doing so has the benefit that the provided password can be a [GitHub personal access token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line).

## Requirements

1. A version of git which supports gitcredentials (which is true for versions >= 1.9.1).

## TODO

1. Support for SSH as prococol.
1. Allow users to specify credentials using environment variables.
1. Make the GIT credential cache timeout configurable via the buildpack.toml to allow for builds which take longer than 15 minutes.
1. More tests are required, particularly for the build phase.

## Authors

* [Khaled Blah](https://github.com/khaledavarteq), [anynines GmbH](https://github.com/anynines)

## LICENSE

MIT, see [LICENSE](./LICENSE)
