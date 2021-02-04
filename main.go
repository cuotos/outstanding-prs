package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

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

	fmt.Printf("Looking for PRs with the following query:\n%s\n\n", queryString)

	prs, err := getPullRequests(client, queryString)
	if err != nil {
		return err
	}

	printOutput(prs, conf.GithubOrg, conf.GithubTeam)

	return nil
}

func getPullRequests(client *github.Client, queryString string) ([]*github.Issue, error) {
	// PRs are "issues" in the world of Github
	var allIssues []*github.Issue

	opt := &github.SearchOptions{
		Sort: "created-desc",
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
	}

	for {
		issues, resp, err := client.Search.Issues(context.Background(), queryString, opt)
		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, issues.Issues...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return allIssues, nil
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

func printOutput(prs []*github.Issue, org, team string) error {

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	// Add headers to the buffer
	writer.Write([]byte("CreatedAt\tTitle\tAuthor\tLink\n"))

	for _, i := range prs {

		formattedIssue := fmt.Sprintf("%s\t%s\t%s\t%s\t\n", i.GetCreatedAt().Format("2006-01-02"), i.GetTitle(), i.GetUser().GetLogin(), i.GetHTMLURL())

		writer.Write([]byte(formattedIssue))
	}

	return writer.Flush()
}

func generateQueryString(org string, members []*github.User) (string, error) {
	queryBuilder := strings.Builder{}

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
