package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/google/go-github/v33/github"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
)

type config struct {
	GithubPat  string `split_words:"true" required:"true"`
	GithubOrg  string `split_words:"true" required:"true"`
	GithubTeam string `split_words:"true" required:"true"`
}

func run() error {

	c := config{}
	err := envconfig.Process("PRS", &c)
	if err != nil {
		return err
	}

	client := getGithubClient(c.GithubPat)

	members, _, err := client.Teams.ListTeamMembersBySlug(context.Background(), c.GithubOrg, c.GithubTeam, nil)

	if err != nil {
		return err
	}

	// PRs _are_ "issues" in the world of Github
	issues, _, err := client.Search.Issues(context.Background(), generateQueryString(members), &github.SearchOptions{Sort: "created-desc"})
	if err != nil {
		return err
	}

	printOutput(issues.Issues, c.GithubOrg, c.GithubTeam)

	return nil
}

func printOutput(prs []*github.Issue, org, team string) error {

	fmt.Printf("Open PRs for %s/%s team\n", org, team)

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	// Add headers to the buffer
	writer.Write([]byte("CreatedAt\tTitle\tAuthor\tLink\n"))

	for _, i := range prs {

		formattedIssue := fmt.Sprintf("%s\t%s\t%s\t%s\t\n", i.GetCreatedAt().Format("2006-01-02"), i.GetTitle(), i.GetUser().GetLogin(), i.GetHTMLURL())
		writer.Write([]byte(formattedIssue))
	}

	return writer.Flush()
}

func generateQueryString(members []*github.User) string {
	usersQueryString := strings.Builder{}

	for _, m := range members {
		usersQueryString.WriteString(fmt.Sprintf("author:%s ", m.GetLogin()))
	}

	queryString := fmt.Sprintf("org:JSainsburyPLC is:open review:required %s", usersQueryString.String())

	return queryString
}

func getGithubClient(accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	})

	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)

	return client
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
