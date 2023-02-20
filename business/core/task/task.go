package task

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jnkroeker/khyme/business/data/store/task"
	"go.uber.org/zap"
)

type Core struct {
	log  *zap.SugaredLogger
	task task.Store
}

func NewCore(log *zap.SugaredLogger, db *sqlx.DB) Core {
	return Core{
		log:  log,
		task: task.NewStore(log, db),
	}
}

func (c Core) Create(ctx context.Context, url url.URL, now time.Time) (task.Task, error) {

	// PERFORM PRE BUSINESS OPERATIONS

	// create the task based on the user input
	newTask, err := service.CreateTask(url)

	res, err := c.task.Create(ctx, newTask, now)

	if err != nil {
		return task.Task{}, fmt.Errorf("create: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS

	return res, nil
}

func (c Core) Delete(ctx context.Context, taskID string) error {

	// PERFORM PRE BUSINESS OPERATIONS

	if err := c.task.Delete(ctx, taskID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS

	return nil
}

func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]task.Task, error) {

	// PERFORM PRE BUSINESS OPERATIONS

	tasks, err := c.task.Query(ctx, pageNumber, rowsPerPage)

	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS

	return tasks, nil
}
