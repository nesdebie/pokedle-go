// Package server exposes the Pokédle HTTP API and serves the static front.
package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"pokedle/internal/dataset"
	"pokedle/internal/game"
	"pokedle/internal/pokeapi"
)

// Config tells the server where to find its files.
type Config struct {
	StaticDir string // e.g. "static"
	EnvFile   string // e.g. ".env" (holds POKEDLE_SECRET)
	DevMode   bool   // cookies expire after 5 minutes instead of midnight
	PokeAPI   string // PokéAPI base URL; empty means the public API
}

type Server struct {
	cfg    Config
	names  *dataset.NameIndex
	data   game.Data
	secret string
	api    *pokeapi.Client
	static http.Handler
}

// New loads every CSV once and returns a ready-to-serve Server.
func New(cfg Config) (*Server, error) {
	names, err := dataset.LoadNames(dataset.NamesFile)
	if err != nil {
		return nil, err
	}
	evolutions, err := dataset.LoadEvolutions(dataset.EvolutionsFile)
	if err != nil {
		return nil, err
	}
	generations, err := dataset.LoadGenerations(dataset.GenerationsFile)
	if err != nil {
		return nil, err
	}

	api := pokeapi.NewClient()
	if cfg.PokeAPI != "" {
		api.BaseURL = cfg.PokeAPI
	}

	return &Server{
		cfg:    cfg,
		names:  names,
		data:   game.Data{Evolutions: evolutions, Generations: generations},
		secret: game.LoadSecret(cfg.EnvFile),
		api:    api,
		static: http.FileServer(http.Dir(cfg.StaticDir)),
	}, nil
}

// Routes returns the HTTP router.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/static/", http.StripPrefix("/static/", s.static))
	mux.HandleFunc("/api/guess", s.handleGuess)
	mux.HandleFunc("/api/today", s.handleToday)
	mux.HandleFunc("/api/hints", s.handleHints)
	mux.HandleFunc("/api/suggest", s.handleSuggest)
	return mux
}

// handleIndex serves index.html on "/" and any other static file otherwise
// (e.g. /images/favicon.ico).
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.static.ServeHTTP(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.cfg.StaticDir, "index.html"))
}

// todayTargetID returns the ID of the Pokémon to guess today.
func (s *Server) todayTargetID() int {
	return s.names.IDAt(s.todayIndex())
}

func (s *Server) todayIndex() int {
	return game.DailyIndex(s.names.Len(), s.secret, time.Now())
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
