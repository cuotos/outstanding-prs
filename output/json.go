package output

import (
	"encoding/json"
	"io"

	"github.com/cuotos/outstanding-prs/model"
)

func PrintOutput(w io.Writer, prs []*model.PullRequest) error {
	return json.NewEncoder(w).Encode(prs)
}
