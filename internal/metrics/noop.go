package metrics

import "context"

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

func (NoopTracker) Close() error {
	return nil
}
