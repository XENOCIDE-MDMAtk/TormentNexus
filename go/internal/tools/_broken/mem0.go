package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// HandleMem0 creates a typed relation between two memories in the Graph memory store.
// Parameters:
// - source_id (string, required)
// - target_id (string, required)
// - rel_type (string, required)
// - weight (float64, optional, default: 1.0)
// - metadata (JSON map, optional)
func HandleMem0(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if GlobalVectorStore == nil || GlobalVectorStore.RelationStore() == nil {
		return err("Graph RelationStore is not initialized")
	}

	sourceID, _ := getString(args, "source_id")
	targetID, _ := getString(args, "target_id")
	relType, _ := getString(args, "rel_type")
	if sourceID == "" || targetID == "" || relType == "" {
		return err("missing required parameter: source_id, target_id, and rel_type are required")
	}

	weightVal, foundWeight := args["weight"]
	weight := 1.0
	if foundWeight {
		if w, ok := weightVal.(float64); ok {
			weight = w
		}
	}

	metadata := map[string]any{}
	if metaVal, foundMeta := args["metadata"]; foundMeta {
		if m, ok := metaVal.(map[string]any); ok {
			metadata = m
		} else if mStr, ok := metaVal.(string); ok {
			_ = json.Unmarshal([]byte(mStr), &metadata)
		}
	}

	e := GlobalVectorStore.RelationStore().AddRelation(ctx, sourceID, targetID, relType, weight, metadata)
	if e != nil {
		return err(fmt.Sprintf("failed to save relation: %v", e))
	}

	return ok(fmt.Sprintf("Successfully created relation: (%s) -[%s, weight=%.2f]-> (%s)", sourceID, relType, weight, targetID))
}

// HandleMem1 traverses the graph memory store starting from a node.
// Parameters:
// - start_id (string, required)
// - depth (int, optional, default: 2)
// - min_weight (float64, optional, default: 0.1)
func HandleMem1(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if GlobalVectorStore == nil || GlobalVectorStore.RelationStore() == nil {
		return err("Graph RelationStore is not initialized")
	}

	startID, _ := getString(args, "start_id")
	if startID == "" {
		return err("missing required parameter: start_id")
	}

	depth := 2
	if dVal, foundDepth := getInt(args, "depth"); foundDepth {
		depth = dVal
	}

	minWeight := 0.1
	if mwVal, foundMW := args["min_weight"]; foundMW {
		if mw, ok := mwVal.(float64); ok {
			minWeight = mw
		}
	}

	nodes, e := GlobalVectorStore.RelationStore().Traverse(ctx, startID, depth, minWeight)
	if e != nil {
		return err(fmt.Sprintf("traversal failed: %v", e))
	}

	data, _ := json.MarshalIndent(nodes, "", "  ")
	return ok(string(data))
}

// HandleMem2 retrieves all relations (inbound and outbound) for a node or get stats.
// Parameters:
// - node_id (string, optional - if empty, returns overall graph stats)
func HandleMem2(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if GlobalVectorStore == nil || GlobalVectorStore.RelationStore() == nil {
		return err("Graph RelationStore is not initialized")
	}

	nodeID, _ := getString(args, "node_id")
	if nodeID == "" {
		// Return relation stats
		stats, e := GlobalVectorStore.RelationStore().GetStats(ctx)
		if e != nil {
			return err(fmt.Sprintf("failed to get graph stats: %v", e))
		}
		data, _ := json.MarshalIndent(stats, "", "  ")
		return ok(string(data))
	}

	outbound, e := GlobalVectorStore.RelationStore().GetRelations(ctx, nodeID)
	if e != nil {
		return err(fmt.Sprintf("failed to get outbound relations: %v", e))
	}

	inbound, e := GlobalVectorStore.RelationStore().GetInbound(ctx, nodeID)
	if e != nil {
		return err(fmt.Sprintf("failed to get inbound relations: %v", e))
	}

	result := map[string]any{
		"node_id":  nodeID,
		"outbound": outbound,
		"inbound":  inbound,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}