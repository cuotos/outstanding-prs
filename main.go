package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cuotos/outstanding-prs/filter"
	"github.com/google/go-github/v33/github"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
)

var (
	// Set default filters, like "review:required"
	defaultFilters = []filter.FilterOpt{
		filter.WithIsNotDraft(),
		filter.WithIsOpen(),
		filter.WithReviewRequired(),
	}
)

var (
	version = ""
	commit  = ""
)

type config struct {
	GithubPat  string `split_words:"true" required:"true" envconfig:"GITHUB_TOKEN"`
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
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	v := flag.Bool("v", false, "prints version")
	flag.Parse()
	if *v {
		fmt.Printf("%s-%s", version, commit)
		os.Exit(0)
	}

	conf := config{}
	err := envconfig.Process("PRS", &conf)
	if err != nil {
		return err
	}

	client := getGithubClient(conf.GithubPat)

	members, err := getOrgTeamMembers(client, conf.GithubOrg, conf.GithubTeam)
	if err != nil {
		return err
	}

	queryString, err := generateQueryString(conf.GithubOrg, members)
	if err != nil {
		return err
	}

	log.Debugf(`Looking for PRs with the following query: "%s"\n`, queryString)

	prs, err := getPullRequests(client, queryString)
	if err != nil {
		return err
	}

	printOutput(prs, conf.GithubOrg, conf.GithubTeam)

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

func printOutput(prs []PullRequest, org, team string) error {

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	// Add headers to the buffer
	writer.Write([]byte("CreatedAt\tTitle\tAuthor\tHead\tBase\tLink\n"))

	for _, pr := range prs {

		formattedIssue := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t\n", pr.CreatedAt.Format("2006-01-02"), pr.Title, pr.Author, pr.Head, pr.Base, pr.Link)

		writer.Write([]byte(formattedIssue))
	}

	return writer.Flush()
}

func generateQueryString(org string, members []*github.User) (string, error) {
	queryBuilder := strings.Builder{}

	queryBuilder.WriteString("type:pr ")

	var users []string

	for _, m := range members {
		users = append(users, m.GetLogin())
	}

	filters := append(defaultFilters, filter.WithOrg(org), filter.WithAuthors(users...))
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
