package hsync

import (
	"context"
	"testing"
	"time"
)

func TestLinkCrawlerManagerRunOnceUpdatesStatus(t *testing.T) {
	dbPath := createLinkCrawlerTestDB(t)
	manager := NewLinkCrawlerManager(dbPath, 10*time.Millisecond, false)

	report, err := manager.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("expected run once to succeed, got %v", err)
	}
	if report.Selected == 0 {
		t.Fatalf("expected pending links to be selected")
	}

	status := manager.Status()
	if status.TotalRuns != 1 {
		t.Fatalf("expected total runs to be 1, got %d", status.TotalRuns)
	}
	if status.TotalProcessed != report.Processed {
		t.Fatalf("expected total processed to equal report processed, got %d vs %d", status.TotalProcessed, report.Processed)
	}
	if status.LastReport == nil || status.LastReport.Succeeded != report.Succeeded {
		t.Fatalf("expected last report to be recorded, got %+v", status.LastReport)
	}
	if status.LastStartedAt == "" || status.LastFinishedAt == "" {
		t.Fatalf("expected timestamps to be populated, got %+v", status)
	}
}

func TestLinkCrawlerManagerStartStop(t *testing.T) {
	dbPath := createLinkCrawlerTestDB(t)
	manager := NewLinkCrawlerManager(dbPath, 5*time.Millisecond, false)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if !manager.Start(ctx) {
		t.Fatalf("expected manager to start")
	}
	if manager.Start(ctx) {
		t.Fatalf("expected second start to be ignored")
	}

	deadline := time.Now().Add(250 * time.Millisecond)
	status := manager.Status()
	for status.TotalRuns == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		status = manager.Status()
	}
	if !status.Running {
		t.Fatalf("expected running status")
	}
	if status.TotalRuns == 0 {
		t.Fatalf("expected at least one run to have occurred")
	}

	if !manager.Stop() {
		t.Fatalf("expected stop to succeed")
	}
	if manager.Stop() {
		t.Fatalf("expected second stop to be ignored")
	}
	if manager.Status().Running {
		t.Fatalf("expected manager to report stopped")
	}
}
