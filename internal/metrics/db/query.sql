-- name: InsertCommandExecution :one
INSERT INTO command_executions (command, command_type, duration_ms, exit_code, flags)
VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: InsertAIInvocation :one
INSERT INTO ai_invocations (command, model, prompt_length, response_length, latency_ms, success, error)
VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetCommandStats :many
SELECT command, COUNT(*) as count, AVG(duration_ms) as avg_duration_ms
FROM command_executions
WHERE datetime(executed_at) >= datetime(?)
GROUP BY command
ORDER BY count DESC;

-- name: GetAIStats :one
SELECT COUNT(*) as total_calls,
       COALESCE(SUM(prompt_length), 0) as total_prompt_tokens,
       COALESCE(SUM(response_length), 0) as total_response_tokens,
       COALESCE(AVG(latency_ms), 0) as avg_latency_ms
FROM ai_invocations
WHERE datetime(created_at) >= datetime(?);

-- name: GetTotalCommands :one
SELECT COUNT(*) as total FROM command_executions WHERE datetime(executed_at) >= datetime(?);

-- name: GetFailureCount :one
SELECT COUNT(*) as failures FROM command_executions WHERE exit_code != 0 AND datetime(executed_at) >= datetime(?);

-- name: GetRecentCommands :many
SELECT * FROM command_executions
ORDER BY executed_at DESC
LIMIT ?;

-- name: GetRecentAIInvocations :many
SELECT * FROM ai_invocations
ORDER BY created_at DESC
LIMIT ?;
