package metrics

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/emiliopalmerini/grimorio/internal/metrics/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed db/migrations/*.sql
var migrationsFS embed.FS

type CommandType string

const (
	Cantrip CommandType = "cantrip"
	Spell   CommandType = "spell"
)

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
	GetSummary(ctx context.Context, since time.Time) (Summary, error)
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

	mu      sync.Mutex
	sqlDB   *sql.DB
	queries *db.Queries
	init    bool
}

func NewSQLiteTracker(dbPath string) *SQLiteTracker {
	return &SQLiteTracker{dbPath: dbPath}
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

	var flagsSQL sql.NullString
	if flags != "" {
		flagsSQL = sql.NullString{String: flags, Valid: true}
	}

	_, err := t.queries.InsertCommandExecution(ctx, db.InsertCommandExecutionParams{
		Command:     command,
		CommandType: string(cmdType),
		DurationMs:  durationMs,
		ExitCode:    int64(exitCode),
		Flags:       flagsSQL,
	})
	if err != nil {
		return fmt.Errorf("insert command execution: %w", err)
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

	var errSQL sql.NullString
	if errMsg != "" {
		errSQL = sql.NullString{String: errMsg, Valid: true}
	}

	_, err := t.queries.InsertAIInvocation(ctx, db.InsertAIInvocationParams{
		Command:        command,
		Model:          model,
		PromptLength:   sql.NullInt64{Int64: int64(promptLen), Valid: true},
		ResponseLength: sql.NullInt64{Int64: int64(responseLen), Valid: true},
		LatencyMs:      sql.NullInt64{Int64: latencyMs, Valid: true},
		Success:        successInt,
		Error:          errSQL,
	})
	if err != nil {
		return fmt.Errorf("insert ai invocation: %w", err)
	}
	return nil
}

func (t *SQLiteTracker) GetSummary(ctx context.Context, since time.Time) (Summary, error) {
	if err := t.ensureInit(ctx); err != nil {
		return Summary{}, err
	}

	sinceStr := since.Format("2006-01-02 15:04:05")

	total, err := t.queries.GetTotalCommands(ctx, sinceStr)
	if err != nil {
		return Summary{}, fmt.Errorf("get total commands: %w", err)
	}

	failures, err := t.queries.GetFailureCount(ctx, sinceStr)
	if err != nil {
		return Summary{}, fmt.Errorf("get failure count: %w", err)
	}

	aiStats, err := t.queries.GetAIStats(ctx, sinceStr)
	if err != nil {
		return Summary{}, fmt.Errorf("get ai stats: %w", err)
	}

	cmdStats, err := t.queries.GetCommandStats(ctx, sinceStr)
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
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sqlDB != nil {
		return t.sqlDB.Close()
	}
	return nil
}
