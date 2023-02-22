package task

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	taskCore "github.com/jnkroeker/khyme/business/core/task"
	"github.com/jnkroeker/khyme/business/sys/database"
	"github.com/jnkroeker/khyme/business/sys/validate"
	"github.com/jnkroeker/khyme/foundation/web"
)

type Handlers struct {
	Task taskCore.Core
}

func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invallid page format [%s]", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid rows format [%s]", rows), http.StatusBadRequest)
	}

	users, err := h.Task.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return fmt.Errorf("unable to query for users: %w", err)
	}

	return web.Respond(ctx, w, users, http.StatusOK)
}

func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var url url.URL
	if err := web.Decode(r, &url); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	task, err := h.Task.Create(ctx, url, v.Now)
	if err != nil {
		return fmt.Errorf("task[%+v]: %w", &task, err)
	}

	return web.Respond(ctx, w, task, http.StatusCreated)
}

func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := web.Param(r, "id")
	if err := h.Task.Delete(ctx, id); err != nil {
		switch validate.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		case database.ErrForbidden:
			return validate.NewRequestError(err, http.StatusForbidden)
		default:
			return fmt.Errorf("ID[%s]: %w", id, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
