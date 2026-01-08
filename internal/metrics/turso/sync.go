package turso

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/emiliopalmerini/grimorio/internal/metrics/db"
)

const batchSize = 100

type Syncer struct {
	client  *Client
	queries *db.Queries
}

func NewSyncer(client *Client, queries *db.Queries) *Syncer {
	return &Syncer{
		client:  client,
		queries: queries,
	}
}

type SyncResult struct {
	CommandsSynced int
	AISynced       int
}

func (s *Syncer) Sync(ctx context.Context) (*SyncResult, error) {
	if err := s.client.InitSchema(ctx); err != nil {
		return nil, fmt.Errorf("init remote schema: %w", err)
	}

	result := &SyncResult{}

	cmdCount, err := s.syncCommandExecutions(ctx)
	if err != nil {
		return nil, fmt.Errorf("sync command executions: %w", err)
	}
	result.CommandsSynced = cmdCount

	aiCount, err := s.syncAIInvocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("sync ai invocations: %w", err)
	}
	result.AISynced = aiCount

	return result, nil
}

func (s *Syncer) syncCommandExecutions(ctx context.Context) (int, error) {
	total := 0

	for {
		records, err := s.queries.GetUnsyncedCommandExecutions(ctx, batchSize)
		if err != nil {
			return total, fmt.Errorf("get unsynced commands: %w", err)
		}

		if len(records) == 0 {
			break
		}

		statements := make([]statement, len(records))
		ids := make([]int64, len(records))

		for i, rec := range records {
			executedAt := nullArg()
			if rec.ExecutedAt.Valid {
				executedAt = textArg(rec.ExecutedAt.Time.Format(time.RFC3339))
			}

			flags := nullArg()
			if rec.Flags.Valid {
				flags = textArg(rec.Flags.String)
			}

			statements[i] = statement{
				SQL: `INSERT INTO command_executions
					(command, command_type, duration_ms, exit_code, flags, executed_at, machine_id, synced)
					VALUES (?, ?, ?, ?, ?, ?, ?, 1)`,
				Args: []argValue{
					textArg(rec.Command),
					textArg(rec.CommandType),
					intArg(rec.DurationMs),
					intArg(rec.ExitCode),
					flags,
					executedAt,
					textArg(rec.MachineID),
				},
			}
			ids[i] = rec.ID
		}

		if _, err := s.client.ExecuteBatch(ctx, statements); err != nil {
			return total, fmt.Errorf("push commands to turso: %w", err)
		}

		if err := s.queries.MarkCommandExecutionsSynced(ctx, ids); err != nil {
			return total, fmt.Errorf("mark commands synced: %w", err)
		}

		total += len(records)

		if len(records) < batchSize {
			break
		}
	}

	return total, nil
}

func (s *Syncer) syncAIInvocations(ctx context.Context) (int, error) {
	total := 0

	for {
		records, err := s.queries.GetUnsyncedAIInvocations(ctx, batchSize)
		if err != nil {
			return total, fmt.Errorf("get unsynced ai invocations: %w", err)
		}

		if len(records) == 0 {
			break
		}

		statements := make([]statement, len(records))
		ids := make([]int64, len(records))

		for i, rec := range records {
			createdAt := nullArg()
			if rec.CreatedAt.Valid {
				createdAt = textArg(rec.CreatedAt.Time.Format(time.RFC3339))
			}

			promptLen := nullArg()
			if rec.PromptLength.Valid {
				promptLen = intArg(rec.PromptLength.Int64)
			}
			responseLen := nullArg()
			if rec.ResponseLength.Valid {
				responseLen = intArg(rec.ResponseLength.Int64)
			}
			latencyMs := nullArg()
			if rec.LatencyMs.Valid {
				latencyMs = intArg(rec.LatencyMs.Int64)
			}

			errMsg := nullArg()
			if rec.Error.Valid {
				errMsg = textArg(rec.Error.String)
			}

			statements[i] = statement{
				SQL: `INSERT INTO ai_invocations
					(command, model, prompt_length, response_length, latency_ms, success, error, created_at, machine_id, synced)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
				Args: []argValue{
					textArg(rec.Command),
					textArg(rec.Model),
					promptLen,
					responseLen,
					latencyMs,
					intArg(rec.Success),
					errMsg,
					createdAt,
					textArg(rec.MachineID),
				},
			}
			ids[i] = rec.ID
		}

		if _, err := s.client.ExecuteBatch(ctx, statements); err != nil {
			return total, fmt.Errorf("push ai invocations to turso: %w", err)
		}

		if err := s.queries.MarkAIInvocationsSynced(ctx, ids); err != nil {
			return total, fmt.Errorf("mark ai invocations synced: %w", err)
		}

		total += len(records)

		if len(records) < batchSize {
			break
		}
	}

	return total, nil
}

type RemoteSummary struct {
	TotalCommands int64
	TotalFailures int64
	MachineStats  []MachineStats
	CommandStats  []CommandStats
	AIStats       AIStats
}

type MachineStats struct {
	MachineID string
	Count     int64
}

type CommandStats struct {
	Command       string
	Count         int64
	AvgDurationMs float64
}

type AIStats struct {
	TotalCalls          int64
	TotalPromptTokens   int64
	TotalResponseTokens int64
	AvgLatencyMs        float64
}

func (s *Syncer) GetRemoteSummary(ctx context.Context, from, to time.Time) (*RemoteSummary, error) {
	fromStr := from.Format("2006-01-02 15:04:05")
	toStr := to.Format("2006-01-02 15:04:05")
	if to.IsZero() {
		toStr = "9999-12-31 23:59:59"
	}

	summary := &RemoteSummary{}

	totalResult, err := s.client.Execute(ctx, `
		SELECT COUNT(*) as total FROM command_executions
		WHERE datetime(executed_at) >= datetime(?) AND datetime(executed_at) <= datetime(?)
	`, fromStr, toStr)
	if err != nil {
		return nil, fmt.Errorf("get total commands: %w", err)
	}
	if len(totalResult.Rows) > 0 && len(totalResult.Rows[0]) > 0 {
		summary.TotalCommands = extractInt(totalResult.Rows[0][0])
	}

	failuresResult, err := s.client.Execute(ctx, `
		SELECT COUNT(*) as failures FROM command_executions
		WHERE exit_code != 0
		AND datetime(executed_at) >= datetime(?) AND datetime(executed_at) <= datetime(?)
	`, fromStr, toStr)
	if err != nil {
		return nil, fmt.Errorf("get failures: %w", err)
	}
	if len(failuresResult.Rows) > 0 && len(failuresResult.Rows[0]) > 0 {
		summary.TotalFailures = extractInt(failuresResult.Rows[0][0])
	}

	machineResult, err := s.client.Execute(ctx, `
		SELECT machine_id, COUNT(*) as count FROM command_executions
		WHERE machine_id != ''
		AND datetime(executed_at) >= datetime(?) AND datetime(executed_at) <= datetime(?)
		GROUP BY machine_id ORDER BY count DESC
	`, fromStr, toStr)
	if err != nil {
		return nil, fmt.Errorf("get machine stats: %w", err)
	}
	for _, row := range machineResult.Rows {
		if len(row) >= 2 {
			summary.MachineStats = append(summary.MachineStats, MachineStats{
				MachineID: extractString(row[0]),
				Count:     extractInt(row[1]),
			})
		}
	}

	cmdResult, err := s.client.Execute(ctx, `
		SELECT command, COUNT(*) as count, AVG(duration_ms) as avg_duration_ms
		FROM command_executions
		WHERE datetime(executed_at) >= datetime(?) AND datetime(executed_at) <= datetime(?)
		GROUP BY command ORDER BY count DESC
	`, fromStr, toStr)
	if err != nil {
		return nil, fmt.Errorf("get command stats: %w", err)
	}
	for _, row := range cmdResult.Rows {
		if len(row) >= 3 {
			summary.CommandStats = append(summary.CommandStats, CommandStats{
				Command:       extractString(row[0]),
				Count:         extractInt(row[1]),
				AvgDurationMs: extractFloat(row[2]),
			})
		}
	}

	aiResult, err := s.client.Execute(ctx, `
		SELECT COUNT(*) as total_calls,
			COALESCE(SUM(prompt_length), 0) as total_prompt_tokens,
			COALESCE(SUM(response_length), 0) as total_response_tokens,
			COALESCE(AVG(latency_ms), 0) as avg_latency_ms
		FROM ai_invocations
		WHERE datetime(created_at) >= datetime(?) AND datetime(created_at) <= datetime(?)
	`, fromStr, toStr)
	if err != nil {
		return nil, fmt.Errorf("get ai stats: %w", err)
	}
	if len(aiResult.Rows) > 0 && len(aiResult.Rows[0]) >= 4 {
		row := aiResult.Rows[0]
		summary.AIStats.TotalCalls = extractInt(row[0])
		summary.AIStats.TotalPromptTokens = extractInt(row[1])
		summary.AIStats.TotalResponseTokens = extractInt(row[2])
		summary.AIStats.AvgLatencyMs = extractFloat(row[3])
	}

	return summary, nil
}

func extractInt(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case map[string]interface{}:
		if s, ok := val["value"].(string); ok {
			i, _ := strconv.ParseInt(s, 10, 64)
			return i
		}
		if f, ok := val["value"].(float64); ok {
			return int64(f)
		}
	}
	return 0
}

func extractFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case map[string]interface{}:
		if s, ok := val["value"].(string); ok {
			f, _ := strconv.ParseFloat(s, 64)
			return f
		}
		if f, ok := val["value"].(float64); ok {
			return f
		}
	}
	return 0
}

func extractString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case map[string]interface{}:
		if s, ok := val["value"].(string); ok {
			return s
		}
	}
	return ""
}
