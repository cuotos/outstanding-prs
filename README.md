# outstanding-prs
CLI to list outstanding PRs that are waiting for reviews raised by the calling user, or members of their team.

Outstanding-prs uses OAuth to authenticate against Github, this token is then stored in your OS's local keychain.  
If you are using it against an Org you are a member off (as defined by the `PRS_GITHUB_ORG` env var, you'll need to allow access to that org as part of the OAuth approval)

## Install 

### Homebrew
```
  brew tap cuotos/tap
  brew install outstanding-prs

  # or upgrade with
  brew upgrade outstanding-prs
```

### Go
`go get github.com/cuotos/outstanding-prs` will download and install and build the lastest code the binary into your $GOBIN

### Docker
> DEPRECATED, as the cli uses the local keychain storage, you cannot access this from inside docker

~~see https://hub.docker.com/r/cuotos/outstanding-prs  
`docker run --rm -ti -e GITHUB_TOKEN=<Personal Access Token> -e PRS_GITHUB_ORG=<org name> -e PRS_GITHUB_TEAM=<team name> cuotos/outstanding-prs`  
If the env vars are already set in your shell, you can shared them with the container directly  
`docker run --rm -ti -e GITHUB_TOKEN -e PRS_GITHUB_ORG -e PRS_GITHUB_TEAM cuotos/outstanding-prs`~~

### Binaries

Can be found here [releases/latest](https://github.com/cuotos/outstanding-prs/releases/latest)

## Run

Export required vars, or set them in you bash_profile etc

```bash
export PRS_GITHUB_ORG=<github org> 
export PRS_GITHUB_TEAM=<github team>

# DEPRECATED - use oauth and secure local storage
export GITHUB_TOKEN=<your github PAT token>
```

To view your own PRs that are waiting for reviews  
`$ outstanding-prs`

To view all PRs for your github team that are waiting for reviews  
`$ outstanding-prs -team`

To viw all PRs __INCLUDING__ those that have been approved but are still open  
`$ outstanding-prs -approved`

## Building and Releasing

Project uses __Go Releaser__

After tagging the latest commit with a semver tag, `goreleaser release` is all that is required to build and upload the binaries to __Github__, __Homebrew__, and __Docker hub__.
