package sync

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func createLinkCrawlerTestDB(t *testing.T) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "metamcp.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE links_backlog (
			uuid TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			normalized_url TEXT,
			title TEXT,
			description TEXT,
			tags TEXT,
			source TEXT,
			is_duplicate INTEGER,
			duplicate_of TEXT,
			research_status TEXT,
			http_status INTEGER,
			page_title TEXT,
			page_description TEXT,
			favicon_url TEXT,
			researched_at INTEGER,
			cluster_id TEXT,
			bobbybookmarks_bookmark_id INTEGER,
			import_session_id INTEGER,
			raw_payload TEXT,
			synced_at INTEGER,
			created_at INTEGER,
			updated_at INTEGER
		);
		INSERT INTO links_backlog (uuid, url, normalized_url, title, description, tags, source, research_status, created_at, updated_at)
		VALUES
			('link-1', 'http://example.test/page', 'http://example.test/page', 'Example', 'desc', '[]', 'manual', 'pending', 1, 1),
			('link-2', 'http://example.test/already-tagged', 'http://example.test/already-tagged', 'Tagged', 'desc', '["existing"]', 'manual', 'pending', 2, 2),
			('link-3', 'http://example.test/done', 'http://example.test/done', 'Done', 'desc', '[]', 'manual', 'done', 3, 3);
	`); err != nil {
		t.Fatalf("failed to create links_backlog schema: %v", err)
	}

	return dbPath
}

func TestCrawlPendingLinksProcessesPendingBacklogEntries(t *testing.T) {
	dbPath := createLinkCrawlerTestDB(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/page":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`
				<html>
				  <head>
				    <title>Go Native Link Crawler</title>
				    <meta name="description" content="Native backlog research in Go">
				    <link rel="icon" href="/favicon.ico">
				  </head>
				  <body>
				    <main>Borg native crawler extracts backlog metadata and tags.</main>
				  </body>
				</html>
			`))
		case "/already-tagged":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><head><title>Existing Tags</title></head><body>Should keep tags.</body></html>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	rewriteDBURLs(t, dbPath, server.URL)

	classifierCalls := 0
	report, err := CrawlPendingLinks(context.Background(), dbPath, LinkCrawlerOptions{
		Limit: 5,
		Classifier: func(ctx context.Context, title, description, content string) ([]string, error) {
			classifierCalls++
			if title != "Go Native Link Crawler" {
				t.Fatalf("unexpected title passed to classifier: %q", title)
			}
			return []string{"go", "crawler", "backlog"}, nil
		},
	})
	if err != nil {
		t.Fatalf("expected crawl to succeed, got error: %v", err)
	}

	if report.Selected != 2 {
		t.Fatalf("expected 2 selected links, got %d", report.Selected)
	}
	if report.Processed != 2 {
		t.Fatalf("expected 2 processed links, got %d", report.Processed)
	}
	if report.Succeeded != 2 {
		t.Fatalf("expected 2 succeeded links, got %d", report.Succeeded)
	}
	if report.Failed != 0 {
		t.Fatalf("expected 0 failed links, got %d", report.Failed)
	}
	if report.Tagged != 1 {
		t.Fatalf("expected 1 newly tagged link, got %d", report.Tagged)
	}
	if classifierCalls != 1 {
		t.Fatalf("expected classifier to be called once, got %d", classifierCalls)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to reopen sqlite db: %v", err)
	}
	defer db.Close()

	var status string
	var pageTitle sql.NullString
	var pageDescription sql.NullString
	var faviconURL sql.NullString
	var tags string
	var httpStatus sql.NullInt64
	if err := db.QueryRow(`SELECT research_status, page_title, page_description, favicon_url, tags, http_status FROM links_backlog WHERE uuid = 'link-1'`).Scan(&status, &pageTitle, &pageDescription, &faviconURL, &tags, &httpStatus); err != nil {
		t.Fatalf("failed to query crawled link: %v", err)
	}
	if status != "done" {
		t.Fatalf("expected link-1 to be done, got %q", status)
	}
	if pageTitle.String != "Go Native Link Crawler" {
		t.Fatalf("unexpected page title: %q", pageTitle.String)
	}
	if pageDescription.String != "Native backlog research in Go" {
		t.Fatalf("unexpected page description: %q", pageDescription.String)
	}
	if faviconURL.String != "/favicon.ico" {
		t.Fatalf("unexpected favicon url: %q", faviconURL.String)
	}
	if tags != `["go","crawler","backlog"]` {
		t.Fatalf("unexpected tags payload: %s", tags)
	}
	if httpStatus.Int64 != 200 {
		t.Fatalf("unexpected http status: %d", httpStatus.Int64)
	}

	if err := db.QueryRow(`SELECT tags FROM links_backlog WHERE uuid = 'link-2'`).Scan(&tags); err != nil {
		t.Fatalf("failed to query tagged link: %v", err)
	}
	if tags != `["existing"]` {
		t.Fatalf("expected existing tags to be preserved, got %s", tags)
	}
}

func TestCrawlPendingLinksMarksFailures(t *testing.T) {
	dbPath := createLinkCrawlerTestDB(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	rewriteDBURLs(t, dbPath, server.URL)

	report, err := CrawlPendingLinks(context.Background(), dbPath, LinkCrawlerOptions{Limit: 1})
	if err != nil {
		t.Fatalf("expected crawl to complete with report, got error: %v", err)
	}
	if report.Processed != 1 || report.Failed != 1 || report.Succeeded != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if len(report.Errors) == 0 {
		t.Fatalf("expected crawl errors to be recorded")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to reopen sqlite db: %v", err)
	}
	defer db.Close()

	var status string
	var httpStatus sql.NullInt64
	if err := db.QueryRow(`SELECT research_status, http_status FROM links_backlog WHERE uuid = 'link-1'`).Scan(&status, &httpStatus); err != nil {
		t.Fatalf("failed to query failed link: %v", err)
	}
	if status != "failed" {
		t.Fatalf("expected failed status, got %q", status)
	}
	if httpStatus.Int64 != 502 {
		t.Fatalf("expected http status 502, got %d", httpStatus.Int64)
	}
}

func rewriteDBURLs(t *testing.T, dbPath string, baseURL string) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db for rewrite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		UPDATE links_backlog SET url = ?, normalized_url = ? WHERE uuid = 'link-1';
		UPDATE links_backlog SET url = ?, normalized_url = ? WHERE uuid = 'link-2';
	`, baseURL+"/page", baseURL+"/page", baseURL+"/already-tagged", baseURL+"/already-tagged"); err != nil {
		t.Fatalf("failed to rewrite test urls: %v", err)
	}
}
