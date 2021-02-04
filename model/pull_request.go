package model

import (
	"time"
)

type PullRequest struct {
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	User      string    `json:"user"`
	HTMLURL   string    `json:"html_url"`
}
