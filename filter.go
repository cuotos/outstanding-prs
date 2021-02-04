package main

import (
	"fmt"
	"strings"
)

// TODO move these to a package

type filterOpt func() (string, error)

type filterString struct {
	s string
}

func getFilterString(opts ...filterOpt) (string, error) {
	sb := strings.Builder{}

	for _, opt := range opts {
		s, err := opt()
		if err != nil {
			return "", err
		}
		sb.WriteString(s)
		sb.WriteString(" ")
	}

	return strings.TrimSpace(sb.String()), nil
}

func withAuthors(authors ...string) filterOpt {
	return func() (string, error) {
		s := strings.Builder{}
		for _, a := range authors {
			if a == "" {
				return "", fmt.Errorf("author cannot be empty")
			}

			s.WriteString(fmt.Sprintf("author:%s ", a))
		}
		return strings.TrimSpace(s.String()), nil
	}
}

func withOrg(org string) filterOpt {
	return func() (string, error) {
		if org == "" {
			return "", fmt.Errorf("org cannot be empty")
		}

		return fmt.Sprintf("org:%s", org), nil
	}
}

func withIsOpen() filterOpt {
	return func() (string, error) {
		return "is:open", nil
	}
}

func withIsClosed() filterOpt {
	return func() (string, error) {
		return "is:closed", nil
	}
}

func withReviewRequired() filterOpt {
	return func() (string, error) {
		return "review:required", nil
	}
}

func withIsDraft() filterOpt {
	return func() (string, error) {
		return "draft:true", nil
	}
}

func withIsNotDraft() filterOpt {
	return func() (string, error) {
		return "draft:false", nil
	}
}
