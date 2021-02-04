package main

import (
	"testing"

	"github.com/google/go-github/v33/github"
	"github.com/stretchr/testify/assert"
)

func TestGenerateQueryString(t *testing.T) {
	tcs := []struct {
		Org      string
		Users    []string
		Expected string
	}{
		{
			"TestOrg",
			[]string{},
			"org:TestOrg is:open review:required draft:false",
		},
		{
			"SecondOrg",
			[]string{"cuotos"},
			"org:SecondOrg is:open review:required draft:false author:cuotos",
		},
		{
			"AnotherOrg",
			[]string{"cuotos", "danyo"},
			"org:AnotherOrg is:open review:required draft:false author:cuotos author:danyo",
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

		actual := generateQueryString(tc.Org, members)
		assert.Equal(t, tc.Expected, actual)
	}
}
