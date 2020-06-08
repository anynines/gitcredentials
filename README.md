# GIT credentials Cloud native Buildpack

GIT credentials is a Cloud Native Buildpack that allows an app developer to supply credentials for GIT repositories which require authentication.

## What it does

A user of this buildpack can supply a file called `buildpack.yml` in the root directory of the application or supply environment variables to specify credentials.

1. If the `gitcredentials.credentials` array is found in `buildpack.yml` or particular environment variables exist, the [GIT credential cache](https://git-scm.com/docs/gitcredentials) will be initialized by this buildpack. The GIT credential cache stores credentials [exclusively in memory](https://git-scm.com/book/en/v2/Git-Tools-Credential-Storage) (and forgets them after a configurable timeout has expired).
1. In addition to that, it sets a [credential context](https://git-scm.com/docs/gitcredentials#_credential_contexts) so that GIT knows which credentials to use for which protocol, host and path.
1. Lastly, it sets [`url.<base>.insteadOf`](https://git-scm.com/docs/git-config#Documentation/git-config.txt-urlltbasegtinsteadOf) to direct GIT to authenticate using HTTPs instead of SSH. Doing so has the benefit that the provided password can be a [GitHub personal access token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line) which supports limiting access to users supplying a personal access token to certain scopes (in particular you can set the scope for the token to "read-only").

## How to use this buildpack

### 1. via buildpack.yml

Create a file called `buildpack.yml` in the root directory of your app and add an array with the following fields:

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

### 2. Environment variables

|  Variable  |  Description  |  Example  |  Required?  |
|------------|---------------|-----------|-------------|
|  `$GIT_CREDENTIALS_USERNAME`  |  The username to use during authentication  |  userA  |  yes  |
|  `$GIT_CREDENTIALS_PASSWORD`  |  The password to use during authentication  |  password  |  yes  |
|  `$GIT_CREDENTIALS_PROTOCOL`  |  The protocol to be specified for GIT credentials  |  https  |  no  |
|  `$GIT_CREDENTIALS_HOST`  |  The host to be specified for GIT credentials  |  github.com  |  no  |
|  `$GIT_CREDENTIALS_PATH`  |  The path to be specified for GIT credentials  |  /foo.git  |  no  |

The environment variable names correspond to the fields available to [git-credential](https://git-scm.com/docs/git-credential). The semantics of the fields are the same.

If a variable is not required and not specified by the user then the value of the corresponding default variable specified in [buildpack.toml](./buildpack.toml) will be used. E.g. `$GIT_CREDENTIALS_PROTOCOL` is set to `https` if the user does not specify `$GIT_CREDENTIALS_PROTOCOL` themselves.

#### NOTE

The variables `$GIT_CREDENTIALS_USERNAME` and `$GIT_CREDENTIALS_PASSWORD` are mandatory and have to be specified by the user.

## How to configure this buildpack

Configuration for this build package can be specfied in [buildpack.toml](./buildpack.toml). The following configuration fields are supported in `[metadata.configuration]`:

|  Name  |  Description  |  Default value  |
|------------|---------------|-----------|
|  default_timeout  |  The default timeout for the GIT credential cache in seconds, set to "3600" to support builds which run longer  |  "3600"  |
|  default_protocol  |  The default protocol. Currently only https is supported. Used as value for `$GIT_CREDENTIALS_PROTOCOL` if the environment variable is not specified  |  "https"  |
|  default_host  |  The default host (don't we all use GitHub ;-)). Used as value for `$GIT_CREDENTIALS_HOST` if the environment variable is not specified |  "github.com"  |
|  default_path  |  The default path. The path variable allows you to authenticate to different repos using different Git credential. Used as value for `$GIT_CREDENTIALS_PATH` if the environment variable is not specified  |  "/"  |

## Requirements

1. A version of git which supports gitcredentials (which is true for versions >= 1.9.1).

## TODO

1. Support for SSH as prococol.
1. More tests are required, particularly for the build phase.

## Authors

* [Khaled Blah](https://github.com/khaledavarteq), [anynines GmbH](https://github.com/anynines)

## LICENSE

MIT, see [LICENSE](./LICENSE)
