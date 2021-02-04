package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	tcs := []struct {
		Input          filterOpt
		ExpectedOutput string
		// If not "", require and error was thrown with this message
		ExpectError string
	}{
		{withAuthors("dan"), "author:dan", ""},
		{withAuthors(""), "", "author cannot be empty"},
		{withAuthors("dan", "rich"), "author:dan author:rich", ""},
		{withOrg("DanOrg"), "org:DanOrg", ""},
		{withOrg(""), "", "org cannot be empty"},
		{withIsOpen(), "is:open", ""},
		{withIsClosed(), "is:closed", ""},
		{withIsNotDraft(), "draft:false", ""},
		{withIsDraft(), "draft:true", ""},
		{withReviewRequired(), "review:required", ""},
	}

	for _, tc := range tcs {

		actual, err := tc.Input()

		if tc.ExpectError != "" {
			require.EqualError(t, err, tc.ExpectError)
		} else {
			require.NoError(t, err)
		}

		assert.Equal(t, tc.ExpectedOutput, actual)
	}
}

func TestGenerateFilterString(t *testing.T) {

	tcs := []struct {
		InputOptions []filterOpt
		Expected     string
	}{
		{
			[]filterOpt{
				withAuthors("dan"),
				withOrg("danorg"),
				withIsOpen(),
			},
			"author:dan org:danorg is:open",
		},
		{
			[]filterOpt{
				withOrg("testorg"),
				withIsOpen(),
				withReviewRequired(),
				withIsNotDraft(),
				withAuthors("cuotos", "another"),
			},
			"org:testorg is:open review:required draft:false author:cuotos author:another",
		},
	}

	for _, tc := range tcs {
		fs, err := getFilterString(tc.InputOptions...)

		if assert.NoError(t, err) {
			assert.Equal(t, tc.Expected, fs)
		}
	}
}
