package turso

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL string
	AuthToken   string
}

func ConfigFromEnv() *Config {
	url := os.Getenv("TURSO_DATABASE_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")
	if url == "" || token == "" {
		return nil
	}
	return &Config{
		DatabaseURL: url,
		AuthToken:   token,
	}
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	authToken  string
}

func NewClient(config *Config) *Client {
	baseURL := config.DatabaseURL
	if strings.HasPrefix(baseURL, "libsql://") {
		baseURL = strings.Replace(baseURL, "libsql://", "https://", 1)
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		authToken:  config.AuthToken,
	}
}

type pipelineRequest struct {
	Requests []stmtRequest `json:"requests"`
}

type stmtRequest struct {
	Type string    `json:"type"`
	Stmt statement `json:"stmt,omitempty"`
}

type statement struct {
	SQL  string        `json:"sql"`
	Args []interface{} `json:"args,omitempty"`
}

type pipelineResponse struct {
	Results []stmtResult `json:"results"`
}

type stmtResult struct {
	Type     string        `json:"type"`
	Response *execResponse `json:"response,omitempty"`
	Error    *errorResult  `json:"error,omitempty"`
}

type execResponse struct {
	Type   string      `json:"type"`
	Result *execResult `json:"result,omitempty"`
}

type execResult struct {
	Cols         []column        `json:"cols"`
	Rows         [][]interface{} `json:"rows"`
	AffectedRows int64           `json:"affected_row_count"`
	LastInsertID string          `json:"last_insert_rowid,omitempty"`
}

type column struct {
	Name string `json:"name"`
	Type string `json:"decltype,omitempty"`
}

type errorResult struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func (c *Client) Execute(ctx context.Context, sql string, args ...interface{}) (*execResult, error) {
	results, err := c.ExecuteBatch(ctx, []statement{{SQL: sql, Args: args}})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results returned")
	}
	return results[0], nil
}

func (c *Client) ExecuteBatch(ctx context.Context, statements []statement) ([]*execResult, error) {
	requests := make([]stmtRequest, len(statements))
	for i, stmt := range statements {
		requests[i] = stmtRequest{
			Type: "execute",
			Stmt: stmt,
		}
	}
	requests = append(requests, stmtRequest{Type: "close"})

	reqBody := pipelineRequest{Requests: requests}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v2/pipeline", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("turso error (status %d): %s", resp.StatusCode, string(body))
	}

	var pipeResp pipelineResponse
	if err := json.Unmarshal(body, &pipeResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	results := make([]*execResult, 0, len(statements))
	for i, res := range pipeResp.Results {
		if i >= len(statements) {
			break
		}
		if res.Error != nil {
			return nil, fmt.Errorf("statement %d error: %s", i, res.Error.Message)
		}
		if res.Response != nil && res.Response.Result != nil {
			results = append(results, res.Response.Result)
		} else {
			results = append(results, &execResult{})
		}
	}

	return results, nil
}

func (c *Client) InitSchema(ctx context.Context) error {
	statements := []statement{
		{SQL: `CREATE TABLE IF NOT EXISTS command_executions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command TEXT NOT NULL,
			command_type TEXT NOT NULL,
			duration_ms INTEGER NOT NULL,
			exit_code INTEGER NOT NULL,
			flags TEXT,
			executed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			machine_id TEXT NOT NULL DEFAULT '',
			synced INTEGER NOT NULL DEFAULT 0
		)`},
		{SQL: `CREATE TABLE IF NOT EXISTS ai_invocations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command TEXT NOT NULL,
			model TEXT NOT NULL,
			prompt_length INTEGER,
			response_length INTEGER,
			latency_ms INTEGER,
			success INTEGER NOT NULL,
			error TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			machine_id TEXT NOT NULL DEFAULT '',
			synced INTEGER NOT NULL DEFAULT 0
		)`},
		{SQL: `CREATE INDEX IF NOT EXISTS idx_executions_command ON command_executions(command)`},
		{SQL: `CREATE INDEX IF NOT EXISTS idx_executions_date ON command_executions(executed_at)`},
		{SQL: `CREATE INDEX IF NOT EXISTS idx_ai_date ON ai_invocations(created_at)`},
		{SQL: `CREATE INDEX IF NOT EXISTS idx_executions_machine ON command_executions(machine_id)`},
		{SQL: `CREATE INDEX IF NOT EXISTS idx_ai_machine ON ai_invocations(machine_id)`},
	}

	_, err := c.ExecuteBatch(ctx, statements)
	return err
}
