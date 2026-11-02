package server

import (
	"encoding/json"
	"net/http"
	"time"

	"pokedle/internal/game"
)

// GET /api/today
func (s *Server) handleToday(w http.ResponseWriter, r *http.Request) {
	idx := s.todayIndex()
	writeJSON(w, map[string]any{
		"date":         game.DayKey(time.Now()),
		"index":        idx,
		"max":          s.names.Len(),
		"remaining":    s.names.Len() - idx - 1,
		"guessCounter": guessCount(r),
	})
}

// POST /api/suggest {"query": "pik"}
func (s *Server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	writeJSON(w, s.names.Suggest(req.Query))
}
