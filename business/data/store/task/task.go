package task

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jnkroeker/khyme/business/sys/database"
	"github.com/jnkroeker/khyme/business/sys/validate"
	"go.uber.org/zap"
)

// Store manages the set of APIs for Task access
type Store struct {
	log *zap.SugaredLogger
	db  sqlx.ExtContext
}

func NewStore(log *zap.SugaredLogger, db *sqlx.DB) Store {
	return Store{
		log: log,
		db:  db,
	}
}

func (s Store) Create(ctx context.Context, nt NewTask, now time.Time) (Task, error) {
	task := Task{
		ID:             validate.GenerateID(),
		DateCreated:    now,
		Version:        nt.Version,
		InputResource:  nt.InputResource,
		OutputResource: nt.OutputResource,
		Hooks:          nt.Hooks,
		ExecutionImage: nt.ExecutionImage,
		Timeout:        nt.Timeout,
	}

	const q = `INSERT INTO tasks
						(task_id, date_created, version, input_url, output_url, hooks, exec_image, timeout)
				VALUES
						(:task_id, :date_created, :version, :input_url, :output_url, :hooks, :exec_image, :timeout)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, task); err != nil {
		return Task{}, fmt.Errorf("inserting task: %w", err)
	}

	return task, nil
}

func (s Store) Delete(ctx context.Context, taskID string) error {
	data := struct {
		TaskID string `db:"task_id"`
	}{
		TaskID: taskID,
	}

	const q = `DELETE FROM tasks WHERE task_id = :task_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}

	return nil
}

func (s Store) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Task, error) {
	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `SELECT * FROM tasks ORDER BY task_id OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var tasks []Task
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, data, &tasks); err != nil {
		if err == database.ErrNotFound {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("selecting tasks: %w", err)
	}

	return tasks, nil
}
