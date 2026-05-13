package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/borghq/borg-go/internal/orchestration"
)

func (s *Server) handlePairSessionRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var req struct {
		Task    string                        `json:"task"`
		Squad   []orchestration.SquadMember   `json:"squad,omitempty"`
		Options map[string]any                `json:"options,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	if req.Squad != nil {
		s.pairOrchestrator.SetupSquad(req.Squad)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	result, err := s.pairOrchestrator.RunTask(ctx, req.Task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error(), "partialResult": result})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    result,
	})
}

func (s *Server) handlePairSessionStatus(w http.ResponseWriter, r *http.Request) {
	s.pairOrchestrator.Mu.RLock()
	defer s.pairOrchestrator.Mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"state":   s.pairOrchestrator.State,
			"task":    s.pairOrchestrator.Task,
			"history": s.pairOrchestrator.History,
			"squad":   s.pairOrchestrator.Squad,
		},
	})
}

func (s *Server) handlePairSessionRotate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	s.pairOrchestrator.RotateRoles()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "squad": s.pairOrchestrator.Squad})
}
