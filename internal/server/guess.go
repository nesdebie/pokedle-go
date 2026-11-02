package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pokedle/internal/game"
	"pokedle/internal/pokeapi"
)

type guessRequest struct {
	Guess string `json:"guess"`
	Lang  string `json:"lang"`
}

type pokemonCard struct {
	ID             int      `json:"id,omitempty"`
	Name           string   `json:"name"`
	Types          []string `json:"types"`
	Height         int      `json:"height"`
	Weight         int      `json:"weight"`
	Sprite         string   `json:"sprite"`
	Position       *int     `json:"position,omitempty"`
	IsFullyEvolved *int     `json:"isFullyEvolved,omitempty"`
}

type guessResponse struct {
	OK           bool         `json:"ok"`
	Error        string       `json:"error,omitempty"`
	Correct      bool         `json:"correct"`
	Guess        *pokemonCard `json:"guess"`
	Hints        *game.Hints  `json:"hints"`
	Reveal       *pokemonCard `json:"reveal,omitempty"`
	GuessCounter int          `json:"guessCounter"`
}

// POST /api/guess {"guess": "Pikachu"}
func (s *Server) handleGuess(w http.ResponseWriter, r *http.Request) {
	if alreadySolved(r) {
		writeJSON(w, guessResponse{Error: "You already found. Try tomorrow!", Correct: true})
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req guessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	guessID, ok := s.names.Lookup(req.Guess)
	if !ok {
		writeJSON(w, guessResponse{Error: "Incorrect Pokémon name"})
		return
	}

	guess, err1 := s.api.Pokemon(guessID)
	target, err2 := s.api.Pokemon(s.todayTargetID())
	if err1 != nil || err2 != nil {
		writeJSON(w, guessResponse{Error: "PokeAPI Error"})
		return
	}

	count := guessCount(r) + 1
	hints := s.data.Compare(guess, target)
	evo := s.data.Evolutions[guess.ID]

	resp := guessResponse{
		OK:      true,
		Correct: guess.ID == target.ID,
		Guess: &pokemonCard{
			Name:           s.names.EnglishName(guessID),
			Types:          []string{hints.Type1, hints.Type2},
			Height:         guess.Height,
			Weight:         guess.Weight,
			Sprite:         guess.Artwork(),
			Position:       &evo.Position,
			IsFullyEvolved: &evo.IsFullyEvolved,
		},
		Hints:        &hints,
		GuessCounter: count,
	}

	if resp.Correct {
		s.setCookie(w, cookieSolved, "true")
		resp.Reveal = newRevealCard(target)
	}
	s.setCookie(w, cookieGuesses, strconv.Itoa(count))

	writeJSON(w, resp)
}

func newRevealCard(p *pokeapi.Pokemon) *pokemonCard {
	t1, t2 := p.TypePair()
	return &pokemonCard{
		ID:     p.ID,
		Name:   p.Name,
		Types:  []string{t1, t2},
		Height: p.Height,
		Weight: p.Weight,
		Sprite: p.Artwork(),
	}
}
