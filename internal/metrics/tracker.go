package metrics

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/emiliopalmerini/grimorio/internal/metrics/db"
	"github.com/emiliopalmerini/grimorio/internal/metrics/turso"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

const (
	tursoSyncTimeout = 3 * time.Second
	maxDateSentinel  = "9999-12-31 23:59:59"
	timestampFormat  = "2006-01-02 15:04:05" // SQLite-compatible datetime format
)

//go:embed db/migrations/*.sql
var migrationsFS embed.FS

// toNullString converts a string to sql.NullString, treating empty strings as NULL.
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// toNullableArg converts a string to an interface{} suitable for Turso queries,
// returning nil for empty strings to represent NULL.
func toNullableArg(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

type CommandType string

const (
	Cantrip CommandType = "cantrip"
	Spell   CommandType = "spell"
)

type Filter struct {
	From    time.Time
	To      time.Time
	Command string
}

type Summary struct {
	TotalCommands       int64
	TotalFailures       int64
	TotalAICalls        int64
	TotalPromptTokens   int64
	TotalResponseTokens int64
	AvgLatencyMs        float64
	CommandStats        []CommandStat
}

type CommandStat struct {
	Command       string
	Count         int64
	AvgDurationMs float64
}

type Tracker interface {
	RecordCommand(ctx context.Context, command string, cmdType CommandType, durationMs int64, exitCode int, flags string) error
	RecordAI(ctx context.Context, command, model string, promptLen, responseLen int, latencyMs int64, success bool, errMsg string) error
	GetSummary(ctx context.Context, filter Filter) (Summary, error)
	Queries(ctx context.Context) (*db.Queries, error)
	Close() error
}

var Default Tracker = &NoopTracker{}

func Track(command string, cmdType CommandType, flags string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start).Milliseconds()
	exitCode := 0
	if err != nil {
		exitCode = 1
	}
	Default.RecordCommand(context.Background(), command, cmdType, duration, exitCode, flags)
	return err
}

type SQLiteTracker struct {
	dbPath string

	mu          sync.Mutex
	sqlDB       *sql.DB
	queries     *db.Queries
	init        bool
	tursoClient *turso.Client
	syncWg      sync.WaitGroup
}

func NewSQLiteTracker(dbPath string) *SQLiteTracker {
	return &SQLiteTracker{dbPath: dbPath}
}

func (t *SQLiteTracker) SetTursoClient(client *turso.Client) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tursoClient = client
}

func (t *SQLiteTracker) getTursoClient() *turso.Client {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.tursoClient
}

// tursoSyncJob represents a sync operation to be executed against Turso.
type tursoSyncJob struct {
	recordType string
	localID    int64
	query      string
	args       []interface{}
	markSynced func(ctx context.Context, ids []int64) error
}

// syncToTurso executes a sync job against Turso and marks the local record as synced.
func (t *SQLiteTracker) syncToTurso(client *turso.Client, job tursoSyncJob) {
	defer t.syncWg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), tursoSyncTimeout)
	defer cancel()

	if _, err := client.Execute(ctx, job.query, job.args...); err != nil {
		log.Printf("metrics: failed to sync %s to Turso: %v", job.recordType, err)
		return
	}

	if err := job.markSynced(ctx, []int64{job.localID}); err != nil {
		log.Printf("metrics: failed to mark %s %d as synced: %v", job.recordType, job.localID, err)
	}
}

func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	dir := filepath.Join(home, ".grimorio")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}
	return filepath.Join(dir, "metrics.db"), nil
}

func (t *SQLiteTracker) ensureInit(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.init {
		return nil
	}

	sqlDB, err := sql.Open("sqlite", t.dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)

	if err := t.runMigrations(sqlDB); err != nil {
		sqlDB.Close()
		return fmt.Errorf("run migrations: %w", err)
	}

	t.sqlDB = sqlDB
	t.queries = db.New(sqlDB)
	t.init = true
	return nil
}

func (t *SQLiteTracker) runMigrations(sqlDB *sql.DB) error {
	sourceDriver, err := iofs.New(migrationsFS, "db/migrations")
	if err != nil {
		return fmt.Errorf("create source driver: %w", err)
	}

	dbDriver, err := sqlite.WithInstance(sqlDB, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("create db driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", dbDriver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

func (t *SQLiteTracker) RecordCommand(ctx context.Context, command string, cmdType CommandType, durationMs int64, exitCode int, flags string) error {
	if err := t.ensureInit(ctx); err != nil {
		return err
	}

	machineID := GetMachineID()
	executedAt := time.Now()

	result, err := t.queries.InsertCommandExecution(ctx, db.InsertCommandExecutionParams{
		Command:     command,
		CommandType: string(cmdType),
		DurationMs:  durationMs,
		ExitCode:    int64(exitCode),
		Flags:       toNullString(flags),
		MachineID:   machineID,
	})
	if err != nil {
		return fmt.Errorf("insert command execution: %w", err)
	}

	if client := t.getTursoClient(); client != nil {
		t.syncWg.Add(1)
		go t.syncToTurso(client, tursoSyncJob{
			recordType: "command execution",
			localID:    result.ID,
			query: `INSERT INTO command_executions (command, command_type, duration_ms, exit_code, flags, executed_at, machine_id, synced)
				 VALUES (?, ?, ?, ?, ?, ?, ?, 1)`,
			args:       []interface{}{command, string(cmdType), durationMs, int64(exitCode), toNullableArg(flags), executedAt.Format(timestampFormat), machineID},
			markSynced: t.queries.MarkCommandExecutionsSynced,
		})
	}

	return nil
}

func (t *SQLiteTracker) RecordAI(ctx context.Context, command, model string, promptLen, responseLen int, latencyMs int64, success bool, errMsg string) error {
	if err := t.ensureInit(ctx); err != nil {
		return err
	}

	var successInt int64
	if success {
		successInt = 1
	}

	machineID := GetMachineID()
	createdAt := time.Now()

	result, err := t.queries.InsertAIInvocation(ctx, db.InsertAIInvocationParams{
		Command:        command,
		Model:          model,
		PromptLength:   sql.NullInt64{Int64: int64(promptLen), Valid: true},
		ResponseLength: sql.NullInt64{Int64: int64(responseLen), Valid: true},
		LatencyMs:      sql.NullInt64{Int64: latencyMs, Valid: true},
		Success:        successInt,
		Error:          toNullString(errMsg),
		MachineID:      machineID,
	})
	if err != nil {
		return fmt.Errorf("insert ai invocation: %w", err)
	}

	if client := t.getTursoClient(); client != nil {
		t.syncWg.Add(1)
		go t.syncToTurso(client, tursoSyncJob{
			recordType: "AI invocation",
			localID:    result.ID,
			query: `INSERT INTO ai_invocations (command, model, prompt_length, response_length, latency_ms, success, error, created_at, machine_id, synced)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
			args:       []interface{}{command, model, promptLen, responseLen, latencyMs, successInt, toNullableArg(errMsg), createdAt.Format(timestampFormat), machineID},
			markSynced: t.queries.MarkAIInvocationsSynced,
		})
	}

	return nil
}

func (t *SQLiteTracker) GetSummary(ctx context.Context, filter Filter) (Summary, error) {
	if err := t.ensureInit(ctx); err != nil {
		return Summary{}, err
	}

	fromStr := filter.From.Format(timestampFormat)
	toStr := filter.To.Format(timestampFormat)
	if filter.To.IsZero() {
		toStr = maxDateSentinel
	}

	total, err := t.queries.GetTotalCommands(ctx, db.GetTotalCommandsParams{
		FromDate:      fromStr,
		ToDate:        toStr,
		CommandFilter: filter.Command,
	})
	if err != nil {
		return Summary{}, fmt.Errorf("get total commands: %w", err)
	}

	failures, err := t.queries.GetFailureCount(ctx, db.GetFailureCountParams{
		FromDate:      fromStr,
		ToDate:        toStr,
		CommandFilter: filter.Command,
	})
	if err != nil {
		return Summary{}, fmt.Errorf("get failure count: %w", err)
	}

	aiStats, err := t.queries.GetAIStats(ctx, db.GetAIStatsParams{
		FromDate: fromStr,
		ToDate:   toStr,
	})
	if err != nil {
		return Summary{}, fmt.Errorf("get ai stats: %w", err)
	}

	cmdStats, err := t.queries.GetCommandStats(ctx, db.GetCommandStatsParams{
		FromDate:      fromStr,
		ToDate:        toStr,
		CommandFilter: filter.Command,
	})
	if err != nil {
		return Summary{}, fmt.Errorf("get command stats: %w", err)
	}

	var commandStats []CommandStat
	for _, cs := range cmdStats {
		avgDur := 0.0
		if cs.AvgDurationMs.Valid {
			avgDur = cs.AvgDurationMs.Float64
		}
		commandStats = append(commandStats, CommandStat{
			Command:       cs.Command,
			Count:         cs.Count,
			AvgDurationMs: avgDur,
		})
	}

	promptTokens, _ := aiStats.TotalPromptTokens.(int64)
	responseTokens, _ := aiStats.TotalResponseTokens.(int64)
	avgLatency, _ := aiStats.AvgLatencyMs.(float64)

	return Summary{
		TotalCommands:       total,
		TotalFailures:       failures,
		TotalAICalls:        aiStats.TotalCalls,
		TotalPromptTokens:   promptTokens,
		TotalResponseTokens: responseTokens,
		AvgLatencyMs:        avgLatency,
		CommandStats:        commandStats,
	}, nil
}

func (t *SQLiteTracker) Close() error {
	// Wait for any pending Turso syncs to complete
	t.syncWg.Wait()

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sqlDB != nil {
		return t.sqlDB.Close()
	}
	return nil
}

func (t *SQLiteTracker) Queries(ctx context.Context) (*db.Queries, error) {
	if err := t.ensureInit(ctx); err != nil {
		return nil, err
	}
	return t.queries, nil
}
