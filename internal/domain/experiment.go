package domain

import "time"

type Experiment struct {
	ID          string
	Name        string
	Description *string
	Hypothesis  *string
	StartedAt   time.Time
	EndedAt     *time.Time
	IsActive    bool
	CreatedAt   time.Time
}
