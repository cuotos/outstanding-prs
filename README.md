# outstanding-prs
CLI to list outstanding PRs that are waiting for reviews raised by the calling user, or members of their team

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
see https://hub.docker.com/r/cuotos/outstanding-prs

`docker run --rm -ti -e GITHUB_TOKEN=<Personal Access Token> -e PRS_GITHUB_ORG=<org name> -e PRS_GITHUB_TEAM=<team name> cuotos/outstanding-prs`

If the env vars are already set in your shell, you can shared them with the container directly

`docker run --rm -ti -e GITHUB_TOKEN -e PRS_GITHUB_ORG -e PRS_GITHUB_TEAM cuotos/outstanding-prs`

### Binaries

Can be found here [releases/latest](https://github.com/cuotos/outstanding-prs/releases/latest)

## Run

Export required vars, or set them in you bash_profile etc

```bash
export GITHUB_TOKEN=<your github PAT token>
export PRS_GITHUB_ORG=<github org> 
export PRS_GITHUB_TEAM=<github team>
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
