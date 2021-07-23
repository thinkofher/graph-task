package tasks

import "time"

type Task struct {
	ID string
	TaskBuilder
}

type TaskBuilder struct {
	Author   string
	Comment  string
	Deadline time.Time
}

type Report struct {
	By string
	At time.Time
}
