package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	tcs := []struct {
		Input          FilterOpt
		ExpectedOutput string
		// If not "", require and error was thrown with this message
		ExpectError string
	}{
		{WithAuthors("dan"), "author:dan", ""},
		{WithAuthors(""), "", "author cannot be empty"},
		{WithAuthors("dan", "rich"), "author:dan author:rich", ""},
		{WithOrg("DanOrg"), "org:DanOrg", ""},
		{WithOrg(""), "", "org cannot be empty"},
		{WithIsOpen(), "is:open", ""},
		{WithIsClosed(), "is:closed", ""},
		{WithIsNotDraft(), "draft:false", ""},
		{WithIsDraft(), "draft:true", ""},
		{WithReviewRequired(), "review:required", ""},
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
		InputOptions []FilterOpt
		Expected     string
	}{
		{
			[]FilterOpt{
				WithAuthors("dan"),
				WithOrg("danorg"),
				WithIsOpen(),
			},
			"author:dan org:danorg is:open",
		},
		{
			[]FilterOpt{
				WithOrg("testorg"),
				WithIsOpen(),
				WithReviewRequired(),
				WithIsNotDraft(),
				WithAuthors("cuotos", "another"),
			},
			"org:testorg is:open review:required draft:false author:cuotos author:another",
		},
	}

	for _, tc := range tcs {
		fs, err := GetFilterString(tc.InputOptions...)

		if assert.NoError(t, err) {
			assert.Equal(t, tc.Expected, fs)
		}
	}
}
