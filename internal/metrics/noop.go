package metrics

import (
	"context"
	"errors"

	"github.com/emiliopalmerini/grimorio/internal/metrics/db"
)

var ErrNoDatabase = errors.New("no database available")

type NoopTracker struct{}

func (NoopTracker) RecordCommand(context.Context, string, CommandType, int64, int, string) error {
	return nil
}

func (NoopTracker) RecordAI(context.Context, string, string, int, int, int64, bool, string) error {
	return nil
}

func (NoopTracker) GetSummary(context.Context, Filter) (Summary, error) {
	return Summary{}, nil
}

func (NoopTracker) Queries(context.Context) (*db.Queries, error) {
	return nil, ErrNoDatabase
}

func (NoopTracker) Close() error {
	return nil
}
