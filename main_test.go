package main

import (
	"strings"
	"testing"

	"github.com/cuotos/outstanding-prs/filter"
	"github.com/google/go-github/v33/github"
	"github.com/stretchr/testify/assert"
)

func TestGenerateQueryString(t *testing.T) {
	tcs := []struct {
		Org         string
		Users       []string
		Expected    []string
		incApproved bool
	}{
		{
			"TestOrg",
			[]string{},
			[]string{"org:TestOrg", "is:open", "review:required", "draft:false", "type:pr"},
			false,
		},
		{
			"SecondOrg",
			[]string{"cuotos"},
			[]string{"org:SecondOrg", "is:open", "review:required", "draft:false", "author:cuotos", "type:pr"},
			false,
		},
		{
			"AnotherOrg",
			[]string{"cuotos", "danyo"},
			[]string{"org:AnotherOrg", "is:open", "review:required", "draft:false", "type:pr", "author:cuotos", "author:danyo"},
			false,
		},
		{
			"TestOrg",
			[]string{},
			[]string{"org:TestOrg", "is:open", "draft:false", "type:pr"},
			true,
		},
	}

	for _, tc := range tcs {

		var members []*github.User

		for _, m := range tc.Users {
			// ptr foo, else all the items in "members" will have the same Login string
			// if you use &m in the append line
			userName := m

			members = append(members, &github.User{Login: &userName})
		}

		actual, _ := generateQueryString(tc.Org, members, filter.WithReviewRequired(!tc.incApproved)) //"inc approved" is "NOT requiring review", therefore have to negate the "approved" flag
		actualFields := strings.Fields(actual)
		assert.ElementsMatch(t, tc.Expected, actualFields)
	}
}
