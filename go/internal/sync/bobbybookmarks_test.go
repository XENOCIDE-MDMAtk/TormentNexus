package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func createLinksBacklogTestDB(t *testing.T, dbPath string) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE links_backlog (
			uuid TEXT,
			url TEXT PRIMARY KEY,
			normalized_url TEXT,
			title TEXT,
			description TEXT,
			tags TEXT,
			source TEXT,
			is_duplicate BOOLEAN,
			duplicate_of TEXT,
			research_status TEXT,
			http_status INTEGER,
			page_title TEXT,
			page_description TEXT,
			favicon_url TEXT,
			cluster_id TEXT,
			bobbybookmarks_bookmark_id INTEGER,
			import_session_id INTEGER,
			synced_at INTEGER,
			created_at INTEGER,
			updated_at INTEGER
		);
	`); err != nil {
		t.Fatalf("failed to create links_backlog schema: %v", err)
	}
}

func TestSyncBobbyBookmarksUpsertsPaginatedResults(t *testing.T) {
	title1 := "Example 1"
	title2 := "Example 2"
	desc := "Description"
	source := "web"
	status := "pending"
	httpStatus := 200

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "1" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []Bookmark{
					{ID: 1, URL: "https://example.com/1", NormalizedURL: "https://example.com/1", Title: &title1, Description: &desc, Tags: []string{"tag1"}, Source: &source, ResearchStatus: &status, HTTPStatus: &httpStatus},
					{ID: 2, URL: "https://example.com/2", NormalizedURL: "https://example.com/2", Title: &title2, Tags: []string{"tag2"}, Source: &source, ResearchStatus: &status},
				},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []Bookmark{}})
	}))
	defer server.Close()

	dbPath := filepath.Join(t.TempDir(), "metamcp.db")
	createLinksBacklogTestDB(t, dbPath)

	report, err := SyncBobbyBookmarks(context.Background(), dbPath, server.URL, 2, false, false)
	if err != nil {
		t.Fatalf("SyncBobbyBookmarks returned error: %v", err)
	}
	if report.Fetched != 2 || report.Upserted != 2 || report.Pages != 1 {
		t.Fatalf("unexpected sync report: %#v", report)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("expected no sync errors, got %#v", report.Errors)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to reopen sqlite db: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM links_backlog`).Scan(&count); err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 rows after sync, got %d", count)
	}
}

func TestSyncBobbyBookmarksReportsDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":`))
	}))
	defer server.Close()

	dbPath := filepath.Join(t.TempDir(), "metamcp.db")
	createLinksBacklogTestDB(t, dbPath)

	report, err := SyncBobbyBookmarks(context.Background(), dbPath, server.URL, 5, false, false)
	if err != nil {
		t.Fatalf("SyncBobbyBookmarks returned error: %v", err)
	}
	if len(report.Errors) == 0 {
		t.Fatalf("expected decode error in report, got %#v", report)
	}
	if report.Fetched != 0 || report.Upserted != 0 {
		t.Fatalf("expected no upserts on decode error, got %#v", report)
	}
}
