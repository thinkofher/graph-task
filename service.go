package tasks

import (
	"context"
	"errors"
	"fmt"

	uuid "github.com/satori/go.uuid"
)

var ErrStorageFail = errors.New("tasks: failed to read data from storage")
var ErrEmptyResult = errors.New("tasks: empty result")

type Storage interface {
	Add(ctx context.Context, t Task) error
	Get(ctx context.Context, id string) (*Task, error)
	Done(ctx context.Context, id string, r Report) error
}

type TaskLister interface {
	All(ctx context.Context) ([]Task, error)
	OfAuthor(ctx context.Context, author string) ([]Task, error)
	DonyBy(ctx context.Context, doer string) ([]Task, error)
}

type Service struct {
	Lister  TaskLister
	Storage Storage
}

func (s *Service) NewTask(ctx context.Context, t TaskBuilder) (string, error) {
	id := uuid.NewV4().String()
	if err := s.Storage.Add(ctx, Task{
		ID:          id,
		TaskBuilder: t,
	}); err != nil {
		return "", fmt.Errorf("s.Storage.Add: %w", err)
	}
	return id, nil
}

func (s *Service) TaskWithID(ctx context.Context, id string) (*Task, error) {
	return s.Storage.Get(ctx, id)
}

func (s *Service) DoneTask(ctx context.Context, id string, r Report) error {
	return s.Storage.Done(ctx, id, r)
}

func (s *Service) AllTasks(ctx context.Context) ([]Task, error) {
	return s.Lister.All(ctx)
}

func (s *Service) AllTasksOfAuthor(ctx context.Context, author string) ([]Task, error) {
	return s.Lister.OfAuthor(ctx, author)
}

func (s *Service) AllTasksDoneBy(ctx context.Context, doer string) ([]Task, error) {
	return s.Lister.DonyBy(ctx, doer)
}
