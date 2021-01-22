# outstanding-prs
CLI to list outstanding PRs that are waiting for reviews raised by members of a team

## Install 

### Go
`go get github.com/cuotos/outstanding-prs` will download and install the binary into your $GOBIN

### Docker
see https://hub.docker.com/r/cuotos/outstanding-prs

`docker run --rm -ti -e PRS_GITHUB_PAT=<Personal Access Token> cuotos/outstanding-prs`

### Binaries

Can be found here [releases/latest](https://github.com/cuotos/outstanding-prs/releases/latest)

## Run

Export required vars, or set them in you bash_profile etc

```bash
export PRS_GITHUB_PAT=<your github PAT token> 
export PRS_GITHUB_ORG=<github org> 
export PRS_GITHUB_TEAM=<github team>

$ outstanding-prs
```