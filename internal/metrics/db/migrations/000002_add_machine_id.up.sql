ALTER TABLE command_executions ADD COLUMN machine_id TEXT NOT NULL DEFAULT '';
ALTER TABLE ai_invocations ADD COLUMN machine_id TEXT NOT NULL DEFAULT '';
ALTER TABLE command_executions ADD COLUMN synced INTEGER NOT NULL DEFAULT 0;
ALTER TABLE ai_invocations ADD COLUMN synced INTEGER NOT NULL DEFAULT 0;

CREATE INDEX idx_executions_machine ON command_executions(machine_id);
CREATE INDEX idx_ai_machine ON ai_invocations(machine_id);
CREATE INDEX idx_executions_synced ON command_executions(synced);
CREATE INDEX idx_ai_synced ON ai_invocations(synced);
