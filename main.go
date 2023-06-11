package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cuotos/outstanding-prs/filter"
	"github.com/google/go-github/v33/github"
	"github.com/hashicorp/logutils"
	"github.com/kelseyhightower/envconfig"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"

	"github.com/cli/oauth"
	"github.com/cli/oauth/device"
)

const (
	clientId = "e95029f474d2dac53e6a"
)

var (
	// Set default filters, like "draft:false"
	defaultFilters = []filter.FilterOpt{
		filter.WithIncludeDraft(false),
		filter.WithIsOpen(),
	}
	oauthScopes = []string{"repo"}
)

var (
	version = "unset"
	commit  = "unset"
)

type config struct {
	// GithubPat  string `split_words:"true" required:"true" envconfig:"GITHUB_TOKEN"`
	GithubOrg  string `split_words:"true" required:"true"`
	GithubTeam string `split_words:"true" required:"true"`
}

type PullRequest struct {
	CreatedAt time.Time
	Title     string
	Author    string
	Head      string
	Base      string
	Link      string
	Draft     bool
}

type flags struct {
	version         *bool
	jsonOutput      *bool
	all             *bool
	incApproved     *bool
	searchWholeTeam *bool
	incDrafts       *bool
	reauth          *bool
	logLevel        *string
}

func main() {
	flags := flags{}
	flags.version = flag.Bool("v", false, "prints version")
	flags.jsonOutput = flag.Bool("json", false, "print output in JSON format")
	flags.all = flag.Bool("all", false, "include PRs ready to merge. DEPRECATED: use -approved")
	flags.incApproved = flag.Bool("approved", false, "include PRs ready to merge")
	flags.searchWholeTeam = flag.Bool("team", false, "should look up all members of the team. defaults to just calling user")
	flags.incDrafts = flag.Bool("drafts", false, "include PRs that are in draft")
	flags.reauth = flag.Bool("reauth", false, "clears tokens and authorise against Github.com")
	flags.logLevel = flag.String("log-level", "INFO", "set log level DEBUG, INFO, WARN, or ERROR")

	flag.Parse()

	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logFilter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(strings.ToUpper(*flags.logLevel)),
		Writer:   os.Stderr,
	}
	log.SetOutput(logFilter)

	if err := run(flags); err != nil {
		log.Printf("[ERROR] %s", err)
	}
}

func run(flags flags) error {

	if *flags.version {
		fmt.Printf("%s-%s", version, commit)
		os.Exit(0)
	}

	conf := config{}
	err := envconfig.Process("PRS", &conf)
	if err != nil {
		return err
	}

	accessToken, err := getGithubOAuthAccessToken(*flags.reauth)
	if err != nil {
		return err
	}

	client := getGithubClient(accessToken)

	members := []*github.User{}
	if *flags.searchWholeTeam {
		members, err = getOrgTeamMembers(client, conf.GithubOrg, conf.GithubTeam)
		if err != nil {
			return err
		}
	} else {
		user, _, err := client.Users.Get(context.Background(), "")
		if err != nil {
			return err
		}
		members = append(members, user)
	}

	queryString, err := generateQueryString(conf.GithubOrg, members, filter.WithIncludeApproved(*flags.all || *flags.incApproved), filter.WithIncludeDraft(*flags.incDrafts))
	if err != nil {
		return err
	}

	log.Printf(`[DEBUG] Looking for PRs with the following query: "%s"`, queryString)

	prs, err := getPullRequests(client, queryString)
	if err != nil {
		return err
	}

	if *flags.jsonOutput {
		printOutputJSON(prs)
	} else {
		printOutput(prs, conf.GithubOrg, conf.GithubTeam, *flags.incDrafts)
	}

	return nil
}

func getPullRequestFromIssue(client *github.Client, issue *github.Issue) (*github.PullRequest, error) {

	u, _ := url.Parse(issue.GetPullRequestLinks().GetHTMLURL())
	// /{org}/{repo}/pulls/{number}
	// but we can get number from the issue itself.
	re := regexp.MustCompile("/(.+?)/(.+?)/pull/.+")
	match := re.FindStringSubmatch(u.Path)

	org := match[1]
	repo := match[2]
	pr, _, err := client.PullRequests.Get(context.Background(), org, repo, issue.GetNumber())
	if err != nil {
		return nil, fmt.Errorf("unable to get PR from Issue: %w", err)
	}
	return pr, nil

}

func getPullRequests(client *github.Client, queryString string) ([]PullRequest, error) {
	var allPrs []PullRequest

	// Github PRs are "Issues" in regards to searching
	// once you have the "issues", convert them to a list of "pulls"
	// There doesnt seem to be an easy way to get a PR from its URL, so have to break the URL
	// up into its fields and do a github.GetPR() call for each one.
	opt := &github.SearchOptions{
		Sort: "created-desc",
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
	}

	for {
		foundIssues, resp, err := client.Search.Issues(context.Background(), queryString, opt)
		if err != nil {
			return nil, err
		}

		for _, issue := range foundIssues.Issues {
			ghPullRequest, err := getPullRequestFromIssue(client, issue)
			if err != nil {
				return nil, fmt.Errorf("unable to get github pr: %w", err)
			}

			pr := PullRequest{
				CreatedAt: issue.GetCreatedAt(),
				Title:     issue.GetTitle(),
				Author:    issue.GetUser().GetLogin(),
				Head:      ghPullRequest.GetHead().GetRef(),
				Base:      ghPullRequest.GetBase().GetRef(),
				Link:      issue.GetHTMLURL(),
				Draft:     ghPullRequest.GetDraft(),
			}

			allPrs = append(allPrs, pr)
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return allPrs, nil
}

func getOrgTeamMembers(client *github.Client, org, team string) ([]*github.User, error) {
	var allMembers []*github.User

	opts := &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	}

	for {
		members, resp, err := client.Teams.ListTeamMembersBySlug(context.Background(), org, team, opts)
		if err != nil {
			return nil, err
		}

		allMembers = append(allMembers, members...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return allMembers, nil
}

func printOutputJSON(prs []PullRequest) error {
	jsonBytes, err := json.Marshal(prs)
	if err != nil {
		return fmt.Errorf("unable to marshal PRs: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

func printOutput(prs []PullRequest, org, team string, addDraftsCol bool) error {

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	// Add headers to the buffer {
	headerBuf := bytes.NewBufferString("CreatedAt\tTitle\tAuthor\tHead (from)\tBase (into)\tLink")

	if addDraftsCol {
		headerBuf.WriteString("\tDraft")
	}

	headerBuf.WriteString("\n")

	writer.Write(headerBuf.Bytes())

	for _, pr := range prs {
		issueBuf := bytes.NewBufferString("")
		issueBuf.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", pr.CreatedAt.Format("2006-01-02"), pr.Title, pr.Author, pr.Head, pr.Base, pr.Link))

		if addDraftsCol {
			issueBuf.WriteString(fmt.Sprintf("\t%t", pr.Draft))
		}

		issueBuf.WriteString("\n")

		writer.Write(issueBuf.Bytes())
	}

	return writer.Flush()
}

func generateQueryString(org string, members []*github.User, additionalFilters ...filter.FilterOpt) (string, error) {
	queryBuilder := strings.Builder{}

	queryBuilder.WriteString("type:pr ")

	var users []string

	for _, m := range members {
		users = append(users, m.GetLogin())
	}

	filters := append(defaultFilters, filter.WithOrg(org), filter.WithAuthors(users...))
	filters = append(filters, additionalFilters...)
	fs, err := filter.GetFilterString(filters...)
	if err != nil {
		return "", err
	}
	queryBuilder.WriteString(fs)

	return queryBuilder.String(), nil
}

func getGithubClient(accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	})

	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)

	return client
}

// getGithubOAuthAccessToken will retrieve the token from the users keychain, if its not found it will prompt to authenticate against github.com
func getGithubOAuthAccessToken(reauth bool) (string, error) {

	var token string

	// check if token file exists
	token, err := keyring.Get("outstanding-prs", "")

	// Token file exists
	if err == nil && !reauth {
		return token, nil

	}

	if err == keyring.ErrNotFound || reauth {
		// Token file does not exists

		log.Println("[WARN] token not found, begin auth flow")

		flow := &oauth.Flow{
			Host:     oauth.GitHubHost("https://github.com"),
			ClientID: clientId,
			Scopes:   oauthScopes,
		}

		fmt.Println(flow.Host.DeviceCodeURL)
		ghToken, err := flow.DeviceFlow()
		if err != nil {
			return token, err
		}
		token = ghToken.Token

		if errors.Is(err, device.ErrUnsupported) {
			return token, errors.New("OAuth device flow is not supported, please contact the developer")
		} else if err != nil {
			return token, err
		}

		// store token
		err = keyring.Set("outstanding-prs", "", token)
		if err != nil {
			return token, err
		}

		return token, nil

	}

	return token, fmt.Errorf("failed to retrieve secret from keystore: %w", err)
}
