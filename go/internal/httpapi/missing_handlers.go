package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/memorystore"
)

func (s *Server) handleGetMemory(w http.ResponseWriter, r *http.Request) {
	s.handleMemoryList(w, r)
}

func (s *Server) handleExecuteCode(w http.ResponseWriter, r *http.Request) {
	s.handleCodeExec(w, r)
}

func (s *Server) handleMemorySearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		query = strings.TrimSpace(r.URL.Query().Get("q"))
	}
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing query parameter",
		})
		return
	}
	limit := 5
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	payload := map[string]any{"query": query, "limit": limit}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.query", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.query",
			},
		})
		return
	}

	results, localErr := s.localMemoryQueryResults(query, limit)
	if localErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   localErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    results,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.query",
			"reason":    "upstream unavailable; using local persisted memory search",
		},
	})
}

func (s *Server) handleMemoryContexts(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.listContexts", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.listContexts",
			},
		})
		return
	}

	contexts, localErr := s.localMemoryContexts()
	if localErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   localErr.Error(),
		})
		return
	}

	// Format contexts for compatibility
	formatted := make([]map[string]any, 0, len(contexts))
	for index, ctx := range contexts {
		metadata, _ := ctx["metadata"].(map[string]any)
		responseMetadata := cloneMap(metadata)
		responseMetadata["title"] = stringValue(ctx["title"])
		responseMetadata["source"] = stringValue(ctx["source"])
		responseMetadata["createdAt"] = ctx["createdAt"]
		responseMetadata["chunks"] = ctx["chunks"]
		formatted = append(formatted, map[string]any{
			"id":       localMemoryContextID(ctx, index+1),
			"content":  stringValue(ctx["content"]),
			"metadata": responseMetadata,
			"score":    1,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    formatted,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.listContexts",
			"reason":    "upstream unavailable; using local persisted contexts",
		},
	})
}

func (s *Server) handleMemorySectionedStatus(w http.ResponseWriter, r *http.Request) {
	status, err := memorystore.ReadStatus(s.cfg.WorkspaceRoot)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    status,
	})
}

func (s *Server) handleMemoryArchiveSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var req struct {
		SessionID string   `json:"sessionId"`
		History   []string `json:"history"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON payload: " + err.Error(),
		})
		return
	}

	if req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing sessionId",
		})
		return
	}

	if s.memoryArchiver == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "memory archiver not initialized",
		})
		return
	}

	err := s.memoryArchiver.TakeSnapshot(r.Context(), req.SessionID, req.History)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to take session snapshot: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}
