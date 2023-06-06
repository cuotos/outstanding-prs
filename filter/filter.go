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

// WithReviewRequired will add the "review:required" flag to the search string. It logically does the opposite of WithIncludeApproved
func WithReviewRequired(required bool) FilterOpt {
	return func() (string, error) {
		if required {
			return "review:required", nil
		} else {
			return "", nil
		}
	}
}

// WithIncludeApproved will NOT add the "review:required" flag to the search string. It logically does the opposite of WithReviewRequired
func WithIncludeApproved(approved bool) FilterOpt {
	return func() (string, error) {
		if approved {
			return "", nil
		} else {
			return "review:required", nil
		}
	}
}

func WithIncludeDraft(draft bool) FilterOpt {
	return func() (string, error) {
		if draft {
			return "draft:true", nil
		} else {
			return "draft:false", nil
		}
	}
}
