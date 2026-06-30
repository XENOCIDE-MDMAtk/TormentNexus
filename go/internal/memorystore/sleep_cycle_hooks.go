package memorystore

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

// buryOrphanedMemories finds memories whose session_id no longer has any
// corresponding session records and buries them in L4 limbo as "lost".
func BuryOrphanedMemories(ctx context.Context, db *sql.DB, limbo *L4LimboVault) error {
	rows, err := db.QueryContext(ctx, `
		SELECT id, session_id, memory_type, memory_kind, category, tags,
		       source_url, content, importance, heat_score, last_accessed_at, created_at
		FROM l2_vault
		WHERE session_id != 'manual'
		  AND session_id NOT IN (
			SELECT DISTINCT session_id FROM imported_session_memories
		  )
		  AND heat_score < 15.0
		LIMIT 50
	`)
	if err != nil {
		return fmt.Errorf("buryOrphanedMemories query: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var r controlplane.L2VaultRecord
		var mType, lastAccessStr, createdStr string
		if err := rows.Scan(&r.ID, &r.SessionID, &mType, &r.Kind, &r.Category,
			&r.Tags, &r.SourceURL, &r.Content, &r.Importance, &r.HeatScore,
			&lastAccessStr, &createdStr); err != nil {
			continue
		}
		r.Type = controlplane.MemoryType(mType)
		if t, err := time.Parse(time.RFC3339, lastAccessStr); err == nil {
			r.LastAccessedAt = t
		}
		if t, err := time.Parse(time.RFC3339, createdStr); err == nil {
			r.CreatedAt = t
		}

		if err := limbo.Bury(ctx, r, LimboLost); err != nil {
			continue
		}
		// Remove from L2
		_, _ = db.ExecContext(ctx, `DELETE FROM l2_vault WHERE id = ?`, r.ID)
		_, _ = db.ExecContext(ctx, `DELETE FROM vec_l2_vault WHERE id = ?`, r.ID)
		_, _ = db.ExecContext(ctx, `DELETE FROM l2_memory_fts WHERE memory_id = ?`, r.ID)
		count++
	}

	if count > 0 {
		fmt.Printf("SleepCycle: buried %d orphaned memories in L4 limbo\n", count)
	}
	return nil
}

// dreamCycle automatically reviews a batch of due memories with a simulated
// "recall" score based on their heat score (warmer = better recall).
// This runs the spaced repetition engine without user interaction.
func DreamCycle(ctx context.Context, db *sql.DB) error {
	// Get memories that are due for review, prioritising highest importance first
	rows, err := db.QueryContext(ctx, `
		SELECT l.id, l.importance, l.heat_score
		FROM l2_vault l
		LEFT JOIN spaced_repetition_metadata s ON l.id = s.memory_id
		WHERE s.memory_id IS NULL
		   OR datetime(s.next_review_at) <= datetime(?)
		ORDER BY l.importance DESC
		LIMIT 20
	`, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("dreamCycle query: %w", err)
	}
	defer rows.Close()

	type dueEntry struct {
		id         string
		importance float64
		heat       float64
	}
	var due []dueEntry
	for rows.Next() {
		var d dueEntry
		if err := rows.Scan(&d.id, &d.importance, &d.heat); err != nil {
			continue
		}
		due = append(due, d)
	}

	if len(due) == 0 {
		return nil
	}

	for _, d := range due {
		// Simulate a quality score based on how important + hot the memory is.
		// Warmer, more important memories get higher simulated recall scores.
		rawScore := (d.importance * 2.0) + (d.heat / 50.0)
		// Clamp to 1-4 range
		quality := int(math.Round(rawScore))
		if quality < 1 {
			quality = 1
		}
		if quality > 4 {
			quality = 4
		}

		// Record the review in spaced repetition
		reviewErr := recordReview(ctx, db, d.id, quality)
		if reviewErr != nil {
			continue
		}

		// Boost heat slightly on successful dream-recall
		_, _ = db.ExecContext(ctx, `
			UPDATE l2_vault SET heat_score = MIN(heat_score + 2.0, 100.0), last_accessed_at = datetime('now')
			WHERE id = ?
		`, d.id)
	}

	fmt.Printf("SleepCycle: dreamed %d memories (auto-reviewed via spaced repetition)\n", len(due))
	return nil
}

// recordReview stores a spaced repetition review result.
func recordReview(ctx context.Context, db *sql.DB, memoryID string, quality int) error {
	// Calculate next interval using SM-2 algorithm
	// Default values for first review
	ef := 2.5 // initial easiness factor
	interval := 1.0

	// Check if there's a previous review
	var prevEF float64
	var prevInterval float64
	var prevRepetitions int
	err := db.QueryRowContext(ctx, `
		SELECT easiness_factor, COALESCE(interval_days, 1.0), repetitions
		FROM spaced_repetition_metadata WHERE memory_id = ?
	`, memoryID).Scan(&prevEF, &prevInterval, &prevRepetitions)
	if err == nil {
		ef = prevEF
		interval = prevInterval
	}

	// SM-2 algorithm
	if quality < 3 {
		// Failed recall — reset interval
		interval = 1.0
		ef = ef - 0.2
	} else {
		if interval == 1.0 {
			interval = 1.0
		} else if interval == 1.0 {
			interval = 6.0
		} else {
			interval = interval * ef
		}
		ef = ef + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
	}

	if ef < 1.3 {
		ef = 1.3
	}

	nextReview := time.Now().Add(time.Duration(interval*24) * time.Hour)

	_, err = db.ExecContext(ctx, `
		INSERT INTO spaced_repetition_metadata (memory_id, easiness_factor, interval_days, repetitions, last_reviewed_at, next_review_at)
		VALUES (?, ?, ?, COALESCE((SELECT repetitions FROM spaced_repetition_metadata WHERE memory_id = ?) + 1, 1), datetime('now'), ?)
		ON CONFLICT(memory_id) DO UPDATE SET
			easiness_factor = excluded.easiness_factor,
			interval_days = excluded.interval_days,
			repetitions = excluded.repetitions,
			last_reviewed_at = datetime('now'),
			next_review_at = excluded.next_review_at
	`, memoryID, ef, interval, memoryID, nextReview.Format(time.RFC3339))
	return err
}

var _ = controlplane.L2VaultRecord{}
var _ = fmt.Sprintf
