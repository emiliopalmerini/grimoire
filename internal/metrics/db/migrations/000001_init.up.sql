CREATE TABLE command_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command TEXT NOT NULL,
    command_type TEXT NOT NULL,
    duration_ms INTEGER NOT NULL,
    exit_code INTEGER NOT NULL,
    flags TEXT,
    executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ai_invocations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command TEXT NOT NULL,
    model TEXT NOT NULL,
    prompt_length INTEGER,
    response_length INTEGER,
    latency_ms INTEGER,
    success INTEGER NOT NULL,
    error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_executions_command ON command_executions(command);
CREATE INDEX idx_executions_date ON command_executions(executed_at);
CREATE INDEX idx_ai_date ON ai_invocations(created_at);
