package sessionimport

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func createSQLiteDB(t *testing.T, path string, schema string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		t.Fatalf("failed to create sqlite schema: %v", err)
	}
	return db
}

func TestBuildPrismDatabaseRecordInputs(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data.db")
	db := createSQLiteDB(t, dbPath, `
		CREATE TABLE session_ledger (
			id TEXT,
			project TEXT,
			summary TEXT,
			event_type TEXT,
			confidence_score REAL,
			importance INTEGER,
			todos TEXT,
			decisions TEXT,
			files_changed TEXT,
			keywords TEXT,
			created_at TEXT,
			role TEXT,
			conversation_id TEXT
		);
		CREATE TABLE session_handoffs (
			project TEXT,
			last_summary TEXT,
			key_context TEXT,
			pending_todo TEXT,
			active_decisions TEXT,
			keywords TEXT,
			active_branch TEXT,
			version INTEGER,
			updated_at TEXT,
			created_at TEXT,
			metadata TEXT
		);
	`)
	defer db.Close()
	if _, err := db.Exec(`
		INSERT INTO session_ledger VALUES (
			'ledger-1','alpha','Summarize the work','correction',0.9,4,
			'["ship feature"]','["prefer go"]','["main.go"]','["go","history"]','2026-04-04T01:00:00Z','architect','conv-1'
		);
		INSERT INTO session_handoffs VALUES (
			'alpha','Last summary','Important context','["follow up"]','["use go"]','["handoff"]','main',2,'2026-04-04T02:00:00Z','2026-04-04T01:30:00Z','{"cwd":"C:/workspace/project"}'
		);
	`); err != nil {
		t.Fatalf("failed to seed prism db: %v", err)
	}

	inputs, err := buildImportedSessionRecordInputsFromDatabase(context.Background(), Candidate{SourceTool: "prism-mcp", SourcePath: dbPath, SessionFormat: "db"})
	if err != nil {
		t.Fatalf("expected prism parsing to succeed, got %v", err)
	}
	if len(inputs) != 2 {
		t.Fatalf("expected two prism record inputs, got %#v", inputs)
	}
	if inputs[0].SessionFormat != "prism-ledger" || inputs[1].SessionFormat != "prism-handoff" {
		t.Fatalf("unexpected prism formats: %#v", inputs)
	}
	if inputs[0].ExternalSessionID == nil || *inputs[0].ExternalSessionID != "conv-1" {
		t.Fatalf("expected ledger external session id, got %#v", inputs[0].ExternalSessionID)
	}
	if inputs[1].WorkingDirectory == nil || *inputs[1].WorkingDirectory != "C:/workspace/project" {
		t.Fatalf("expected handoff working directory, got %#v", inputs[1].WorkingDirectory)
	}
}

func TestBuildLLMDatabaseRecordInputs(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "logs.db")
	db := createSQLiteDB(t, dbPath, `
		CREATE TABLE conversations (id TEXT, name TEXT, model TEXT);
		CREATE TABLE responses (
			id TEXT,
			conversation_id TEXT,
			model TEXT,
			prompt TEXT,
			system TEXT,
			prompt_json TEXT,
			response TEXT,
			response_json TEXT,
			datetime_utc TEXT,
			input_tokens INTEGER,
			output_tokens INTEGER,
			resolved_model TEXT
		);
		CREATE TABLE tool_calls (id INTEGER PRIMARY KEY, response_id TEXT, name TEXT);
		CREATE TABLE tool_results (id INTEGER PRIMARY KEY, response_id TEXT, name TEXT);
	`)
	defer db.Close()
	if _, err := db.Exec(`
		INSERT INTO conversations VALUES ('conv-1','LLM Session','gpt-4');
		INSERT INTO responses VALUES ('resp-1','conv-1','gpt-4','How should we route providers?','System prompt',NULL,'Use openrouter free defaults.',NULL,'2026-04-04T01:00:00Z',11,22,'gpt-4');
		INSERT INTO responses VALUES ('resp-2',NULL,'gpt-4','Standalone prompt',NULL,NULL,'Standalone response',NULL,'2026-04-04T02:00:00Z',3,4,'gpt-4');
		INSERT INTO tool_calls (response_id, name) VALUES ('resp-1','search_tools');
		INSERT INTO tool_results (response_id, name) VALUES ('resp-1','search_tools');
	`); err != nil {
		t.Fatalf("failed to seed llm db: %v", err)
	}

	inputs, err := buildImportedSessionRecordInputsFromDatabase(context.Background(), Candidate{SourceTool: "llm-cli", SourcePath: dbPath, SessionFormat: "db"})
	if err != nil {
		t.Fatalf("expected llm parsing to succeed, got %v", err)
	}
	if len(inputs) != 2 {
		t.Fatalf("expected conversation + orphan response, got %#v", inputs)
	}
	if inputs[0].SessionFormat != "llm-conversation" {
		t.Fatalf("unexpected first llm format: %#v", inputs[0])
	}
	if inputs[1].SessionFormat != "llm-response" {
		t.Fatalf("unexpected second llm format: %#v", inputs[1])
	}
	if !strings.Contains(inputs[0].Transcript, "[Tool Call: search_tools]") || !strings.Contains(inputs[0].Transcript, "[Tool Result: search_tools]") {
		t.Fatalf("expected tool call/result stitching in transcript, got %q", inputs[0].Transcript)
	}
}
