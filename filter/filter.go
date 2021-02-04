package filter

import (
	"fmt"
	"strings"
)

// TODO move these to a package

type FilterOpt func() (string, error)

func GetFilterString(opts ...FilterOpt) (string, error) {
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

func WithAuthors(authors ...string) FilterOpt {
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

func WithOrg(org string) FilterOpt {
	return func() (string, error) {
		if org == "" {
			return "", fmt.Errorf("org cannot be empty")
		}

		return fmt.Sprintf("org:%s", org), nil
	}
}

func WithIsOpen() FilterOpt {
	return func() (string, error) {
		return "is:open", nil
	}
}

func WithIsClosed() FilterOpt {
	return func() (string, error) {
		return "is:closed", nil
	}
}

func WithReviewRequired() FilterOpt {
	return func() (string, error) {
		return "review:required", nil
	}
}

func WithIsDraft() FilterOpt {
	return func() (string, error) {
		return "draft:true", nil
	}
}

func WithIsNotDraft() FilterOpt {
	return func() (string, error) {
		return "draft:false", nil
	}
}
