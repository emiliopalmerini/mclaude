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
	ModelID     *string
	PlanType    *string
	Notes       *string
}

type ExperimentVariable struct {
	ID           int64
	ExperimentID string
	Key          string
	Value        string
}
