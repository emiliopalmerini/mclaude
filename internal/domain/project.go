package domain

import "time"

type Project struct {
	ID        string // SHA256 hash of path
	Path      string
	Name      string
	CreatedAt time.Time
}
