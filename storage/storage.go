package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
	tasks "github.com/thinkofher/graph-task"
)

type Storage struct {
	conn redis.Conn
}

func New(conn redis.Conn) *Storage {
	return &Storage{
		conn: conn,
	}
}

const (
	taskGraphID           = "tasks"
	tasksLabel            = "Task"
	tasksIDProperty       = "taskID"
	tasksAuthorProperty   = "author"
	tasksCommentProperty  = "comment"
	tasksDeadlineProperty = "deadline"
	reportsLabel          = "Report"
)

func (s *Storage) taskGraph() *rg.Graph {
	res := rg.GraphNew(taskGraphID, s.conn)
	return &res
}

func (s *Storage) Add(ctx context.Context, t tasks.Task) error {
	graph := s.taskGraph()
	graph.AddNode(&rg.Node{
		Label: tasksLabel,
		Properties: map[string]interface{}{
			tasksIDProperty:       t.ID,
			tasksAuthorProperty:   t.Author,
			tasksCommentProperty:  t.Comment,
			tasksDeadlineProperty: int(t.Deadline.Unix()),
		},
	})
	if _, err := graph.Commit(); err != nil {
		return fmt.Errorf("graph.Commit: %w", err)
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, id string) (*tasks.Task, error) {
	query := `MATCH (t:Task)
	          WHERE t.taskID = $task_id
	          RETURN t.taskID, t.author, t.comment, t.deadline`
	queryResult, err := s.taskGraph().ParameterizedQuery(query, map[string]interface{}{
		"task_id": id,
	})
	if err != nil {
		return nil, fmt.Errorf("graph.Query: %w", err)
	}

	if queryResult.Empty() {
		return nil, fmt.Errorf("%w: there is no task with id=%s", err, id)
	}

	result := tasks.Task{}
	queryResult.Next()
	r := queryResult.Record()

	rawTaskID, ok := r.Get("t.taskID")
	if !ok {
		return nil, fmt.Errorf("r.Get(t.taskID): %w", tasks.ErrStorageFail)
	}
	taskID, ok := rawTaskID.(string)
	if !ok {
		return nil, fmt.Errorf("rawTaskID.(string): %w", tasks.ErrStorageFail)
	}

	rawAuthor, ok := r.Get("t.author")
	if !ok {
		return nil, fmt.Errorf("r.Get(t.author): %w", tasks.ErrStorageFail)
	}
	author, ok := rawAuthor.(string)
	if !ok {
		return nil, fmt.Errorf("rawAuthor.(string): %w", tasks.ErrStorageFail)
	}

	rawComment, ok := r.Get("t.comment")
	if !ok {
		return nil, fmt.Errorf("r.Get(t.comment): %w", tasks.ErrStorageFail)
	}
	comment, ok := rawComment.(string)
	if !ok {
		return nil, fmt.Errorf("rawComment.(string): %w", tasks.ErrStorageFail)
	}

	rawDeadline, ok := r.Get("t.deadline")
	if !ok {
		return nil, fmt.Errorf("r.Get(t.deadline): %w", tasks.ErrStorageFail)
	}

	deadline, ok := rawDeadline.(int)
	if err != nil {
		return nil, fmt.Errorf("rawDeadline.(int): %w", tasks.ErrStorageFail)
	}

	result = tasks.Task{
		ID: taskID,
		TaskBuilder: tasks.TaskBuilder{
			Author:   author,
			Comment:  comment,
			Deadline: time.Unix(int64(deadline), 0),
		},
	}

	return &result, nil

	return nil, nil
}

func (s *Storage) Done(ctx context.Context, id string, r tasks.Report) error {
	query := `MATCH (t:Task)
	          WHERE t.taskID = $task_id
			  CREATE (t)-[:DONE]->(:Report {by: $task_doer, at: $now})`
	if _, err := s.taskGraph().ParameterizedQuery(query, map[string]interface{}{
		"task_id":   id,
		"task_doer": r.By,
		"now":       int(r.At.Unix()),
	}); err != nil {
		return fmt.Errorf("graph.ParameterizedQuery: %w", err)
	}
	return nil
}

func (s *Storage) All(ctx context.Context) ([]tasks.Task, error) {
	query := `MATCH (t:Task)
	          RETURN t.taskID, t.author, t.comment, t.deadline`
	queryResult, err := s.taskGraph().Query(query)
	if err != nil {
		return nil, fmt.Errorf("graph.Query: %w", err)
	}

	result := []tasks.Task{}
	for queryResult.Next() {
		r := queryResult.Record()

		rawTaskID, ok := r.Get("t.taskID")
		if !ok {
			return nil, fmt.Errorf("r.Get(t.taskID): %w", tasks.ErrStorageFail)
		}
		taskID, ok := rawTaskID.(string)
		if !ok {
			return nil, fmt.Errorf("rawTaskID.(string): %w", tasks.ErrStorageFail)
		}

		rawAuthor, ok := r.Get("t.author")
		if !ok {
			return nil, fmt.Errorf("r.Get(t.author): %w", tasks.ErrStorageFail)
		}
		author, ok := rawAuthor.(string)
		if !ok {
			return nil, fmt.Errorf("rawAuthor.(string): %w", tasks.ErrStorageFail)
		}

		rawComment, ok := r.Get("t.comment")
		if !ok {
			return nil, fmt.Errorf("r.Get(t.comment): %w", tasks.ErrStorageFail)
		}
		comment, ok := rawComment.(string)
		if !ok {
			return nil, fmt.Errorf("rawComment.(string): %w", tasks.ErrStorageFail)
		}

		rawDeadline, ok := r.Get("t.deadline")
		if !ok {
			return nil, fmt.Errorf("r.Get(t.deadline): %w", tasks.ErrStorageFail)
		}

		deadline, ok := rawDeadline.(int)
		if err != nil {
			return nil, fmt.Errorf("rawDeadline.(int): %w", tasks.ErrStorageFail)
		}

		result = append(result, tasks.Task{
			ID: taskID,
			TaskBuilder: tasks.TaskBuilder{
				Author:   author,
				Comment:  comment,
				Deadline: time.Unix(int64(deadline), 0),
			},
		})
	}

	return result, nil
}

func (s *Storage) OfAuthor(ctx context.Context, author string) ([]tasks.Task, error) {
	return nil, nil
}

func (s *Storage) DonyBy(ctx context.Context, doer string) ([]tasks.Task, error) {
	return nil, nil
}
