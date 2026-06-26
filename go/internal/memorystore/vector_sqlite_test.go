package memorystore

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

func TestVectorStoreAdvancedFeatures(t *testing.T) {
	vs, err := NewVectorStore(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer vs.Close()

	ctx := context.Background()

	// 1. Commit standard record with advanced metadata
	record := controlplane.L2VaultRecord{
		ID:             "test-1",
		SessionID:      "session-abc",
		Type:           controlplane.MemoryWorking,
		Kind:           "preference",
		Category:       "coding-style",
		Tags:           "go,clean-code",
		SourceURL:      "http://example.com",
		Content:        "User prefers Go over Python for low latency tasks",
		Importance:     0.8,
		HeatScore:      50.0,
		Embedding:      []float32{0.1, -0.2, 0.3},
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}

	if err := vs.Commit(ctx, record); err != nil {
		t.Fatalf("failed to commit record: %v", err)
	}

	// Commit a second record with different metadata
	record2 := controlplane.L2VaultRecord{
		ID:             "test-2",
		SessionID:      "session-abc",
		Type:           controlplane.MemoryWorking,
		Kind:           "fact",
		Category:       "project-info",
		Tags:           "tormentnexus",
		SourceURL:      "http://example.com/project",
		Content:        "TormentNexus has 232 Go files",
		Importance:     0.5,
		HeatScore:      30.0,
		Embedding:      []float32{-0.1, 0.4, 0.1},
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}

	if err := vs.Commit(ctx, record2); err != nil {
		t.Fatalf("failed to commit record 2: %v", err)
	}

	// 2. Query with structured search payload (category filter)
	payload := QueryPayload{
		QueryText: "prefers",
		Category:  "coding-style",
	}
	payloadBytes, _ := json.Marshal(payload)

	results, err := vs.SemanticSearch(ctx, string(payloadBytes), 10)
	if err != nil {
		t.Fatalf("failed structured search: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	} else {
		res := results[0]
		if res.ID != "test-1" {
			t.Errorf("expected ID 'test-1', got %s", res.ID)
		}
		if res.Kind != "preference" || res.Category != "coding-style" {
			t.Errorf("incorrect metadata fields on retrieved record: kind=%s, category=%s", res.Kind, res.Category)
		}
	}

	// 3. Test structured search by vector and kind filter
	payloadVec := QueryPayload{
		QueryVec: []float32{0.1, -0.25, 0.28}, // close to test-1 embedding
		Kind:     "preference",
	}
	payloadVecBytes, _ := json.Marshal(payloadVec)

	resultsVec, err := vs.SemanticSearch(ctx, string(payloadVecBytes), 10)
	if err != nil {
		t.Fatalf("failed structured vector search: %v", err)
	}

	if len(resultsVec) != 1 {
		t.Errorf("expected 1 vector result, got %d", len(resultsVec))
	} else if resultsVec[0].ID != "test-1" {
		t.Errorf("expected ID 'test-1' from vector search, got %s", resultsVec[0].ID)
	}

	// 4. Test ReinforceMemory
	// Score before: importance=0.8, heat=50.0. After success: importance=0.9, heat=65.0 (since s.incrementHeatLocked adds +10 during search retrieve, wait, incrementHeatLocked was executed during search, so heat score increased. Let's fetch current heat score first).
	var origHeat, origImportance float64
	err = vs.db.QueryRowContext(ctx, "SELECT heat_score, importance FROM l2_vault WHERE id = ?", "test-1").Scan(&origHeat, &origImportance)
	if err != nil {
		t.Fatalf("failed to query test-1 details: %v", err)
	}

	if err := vs.ReinforceMemory(ctx, "test-1", true); err != nil {
		t.Fatalf("failed to reinforce memory: %v", err)
	}

	var newHeat, newImportance float64
	err = vs.db.QueryRowContext(ctx, "SELECT heat_score, importance FROM l2_vault WHERE id = ?", "test-1").Scan(&newHeat, &newImportance)
	if err != nil {
		t.Fatalf("failed to query reinforced test-1 details: %v", err)
	}

	expectedHeat := origHeat + 15.0
	expectedImportance := origImportance + 0.1

	if newHeat != expectedHeat {
		t.Errorf("expected heat %f, got %f", expectedHeat, newHeat)
	}
	if mathAbs(newImportance-expectedImportance) > 0.0001 {
		t.Errorf("expected importance %f, got %f", expectedImportance, newImportance)
	}

	// 5. Test GraphRAG relations
	if err := vs.AddRelation(ctx, "test-1", "test-2", "related_to", 0.95); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	rels, err := vs.GetRelations(ctx, "test-1")
	if err != nil {
		t.Fatalf("failed to get relations: %v", err)
	}
	if len(rels) != 1 {
		t.Errorf("expected 1 relation, got %d", len(rels))
	} else {
		if rels[0].SourceID != "test-1" || rels[0].TargetID != "test-2" || rels[0].RelationType != "related_to" || rels[0].Weight != 0.95 {
			t.Errorf("incorrect relation details: %+v", rels[0])
		}
	}

	// 6. Test Forgetting Curve Decay
	if err := vs.ForgettingCurveDecay(ctx); err != nil {
		t.Fatalf("failed forgetting curve decay: %v", err)
	}

	// 7. Test Semantic Consolidation
	// Insert duplicate memory that is semantically similar to test-2
	record3 := controlplane.L2VaultRecord{
		ID:             "test-3",
		SessionID:      "session-abc",
		Type:           controlplane.MemoryWorking,
		Kind:           "fact",
		Category:       "project-info",
		Tags:           "tormentnexus",
		Content:        "TormentNexus has 232 Go files", // Identical to test-2 content to guarantee Jaccard sim > 0.8
		Importance:     0.4,
		HeatScore:      20.0,
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}
	if err := vs.Commit(ctx, record3); err != nil {
		t.Fatalf("failed to commit record 3: %v", err)
	}

	if err := vs.ConsolidateMemories(ctx); err != nil {
		t.Fatalf("failed consolidation: %v", err)
	}

	// test-3 should be merged into test-2 or vice versa, resulting in deletion of test-3
	var count int
	err = vs.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM l2_vault WHERE id = 'test-3'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query test-3 presence: %v", err)
	}
	if count != 0 {
		t.Errorf("expected test-3 to be consolidated/deleted, but it still exists")
	}
}

func TestL3ColdArchiveIntegration(t *testing.T) {
	vs, err := NewVectorStore(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer vs.Close()

	if vs.coldArchive == nil {
		t.Fatalf("expected coldArchive to be initialized, but it was nil")
	}

	ctx := context.Background()

	// 1. Commit active memories (one hot, one cold-candidate)
	hotRecord := controlplane.L2VaultRecord{
		ID:             "hot-1",
		SessionID:      "session-x",
		Type:           controlplane.MemoryWorking,
		Kind:           "fact",
		Category:       "knowledge",
		Tags:           "hot",
		Content:        "This memory has very high relevance and heat score",
		Importance:     0.9,
		HeatScore:      80.0,
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}
	coldCandidate := controlplane.L2VaultRecord{
		ID:             "cold-candidate-1",
		SessionID:      "session-x",
		Type:           controlplane.MemoryWorking,
		Kind:           "fact",
		Category:       "knowledge",
		Tags:           "cold",
		Content:        "This memory has decayed significantly and has very low heat score",
		Importance:     0.4,
		HeatScore:      5.0, // < 10.0 so should be archived on decay
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
	}

	if err := vs.Commit(ctx, hotRecord); err != nil {
		t.Fatalf("failed to commit hot record: %v", err)
	}
	if err := vs.Commit(ctx, coldCandidate); err != nil {
		t.Fatalf("failed to commit cold candidate: %v", err)
	}

	// 2. Run decay process
	if err := vs.ForgettingCurveDecay(ctx); err != nil {
		t.Fatalf("failed to run forgetting curve decay: %v", err)
	}

	// 3. Verify cold candidate has been removed from active l2_vault but exists in l3_cold_archive
	var countL2 int
	err = vs.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM l2_vault WHERE id = 'cold-candidate-1'").Scan(&countL2)
	if err != nil {
		t.Fatalf("failed querying active L2: %v", err)
	}
	if countL2 != 0 {
		t.Errorf("expected cold candidate to be removed from L2, but found %d matches", countL2)
	}

	// Verify hot record is still in L2
	err = vs.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM l2_vault WHERE id = 'hot-1'").Scan(&countL2)
	if err != nil {
		t.Fatalf("failed querying active L2 for hot record: %v", err)
	}
	if countL2 != 1 {
		t.Errorf("expected hot record to remain in L2, but found %d matches", countL2)
	}

	// Check cold archive counts
	countL3, err := vs.coldArchive.Count(ctx)
	if err != nil {
		t.Fatalf("failed to count cold archive records: %v", err)
	}
	if countL3 != 1 {
		t.Errorf("expected 1 record in L3 cold archive, got %d", countL3)
	}

	// 4. Query L3 fallback search
	// Create query text payload that matches the cold archive content keyword 'decayed'
	payload := QueryPayload{
		QueryText: "decayed",
	}
	payloadBytes, _ := json.Marshal(payload)

	results, err := vs.SemanticSearch(ctx, string(payloadBytes), 5)
	if err != nil {
		t.Fatalf("fallback search failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result from fallback search, got %d", len(results))
	}

	if results[0].ID != "cold-candidate-1" {
		t.Errorf("expected result to be 'cold-candidate-1', got %s", results[0].ID)
	}

	// 5. Verify promotion: promoted record should be back in L2 with heat_score = 25.0 and deleted from L3

	err = vs.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM l2_vault WHERE id = 'cold-candidate-1'").Scan(&countL2)
	if err != nil {
		t.Fatalf("failed querying active L2 after promotion: %v", err)
	}
	if countL2 != 1 {
		t.Errorf("expected cold candidate to be promoted back to L2, but it was not found")
	}

	var promotedHeat float64
	var promotedType string
	err = vs.db.QueryRowContext(ctx, "SELECT heat_score, memory_type FROM l2_vault WHERE id = 'cold-candidate-1'").Scan(&promotedHeat, &promotedType)
	if err != nil {
		t.Fatalf("failed querying details of promoted record: %v", err)
	}
	if promotedHeat != 25.0 {
		t.Errorf("expected promoted heat score to be 25.0, got %f", promotedHeat)
	}
	if promotedType != string(controlplane.MemoryLongTerm) {
		t.Errorf("expected promoted memory type to be %s, got %s", controlplane.MemoryLongTerm, promotedType)
	}

	// Verify deleted from L3
	countL3, err = vs.coldArchive.Count(ctx)
	if err != nil {
		t.Fatalf("failed to count cold archive records after promotion: %v", err)
	}
	if countL3 != 0 {
		t.Errorf("expected L3 cold archive to be empty after promotion, got %d records", countL3)
	}
}

func mathAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

