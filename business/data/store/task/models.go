package task

import (
	"net/url"
	"time"
)

// Task represents a data processing task to be executed
type Task struct {
	ID             string        `db:"task_id" json:"id"`
	DateCreated    time.Time     `db:"date_created" json:"date_created"`
	Version        string        `db:"version" json:"version"`
	InputResource  url.URL       `db:"input_url" json:"input_url"`
	OutputResource url.URL       `db:"output_url" json:"output_url"`
	Hooks          string        `db:"hooks" json:"hooks"`
	ExecutionImage string        `db:"exec_image" json:"exec_image"`
	Timeout        time.Duration `db:"timeout" json:"timeout"`
}

// NewTask contains information needed to create a new Task
/*
 * what are we actually accepting from the user?
 * eventually its a video file(s) that we upload to a bucket
 * right now lets assume we are receiving an GCP object path
 */
type NewTask struct {
	Version        string        `db:"version" json:"version"`
	InputResource  url.URL       `db:"input_url" json:"input_url"`
	OutputResource url.URL       `db:"output_url" json:"output_url"`
	Hooks          string        `db:"hooks" json:"hooks"`
	ExecutionImage string        `db:"exec_image" json:"exec_image"`
	Timeout        time.Duration `db:"timeout" json:"timeout"`
}
