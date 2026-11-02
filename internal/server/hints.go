package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pokedle/internal/dataset"
)

// A new hint tier is unlocked every guessesPerTier guesses:
//
//	tier 1: the Pokémon cry
//	tier 2: its types
//	tier 3: its Pokédex description (every available language)
const guessesPerTier = 3

type hintsResponse struct {
	Tier        int               `json:"tier"`
	Cry         string            `json:"cry,omitempty"`
	Types       []string          `json:"types,omitempty"`
	Description map[string]string `json:"description,omitempty"`
}

// GET /api/hints
func (s *Server) handleHints(w http.ResponseWriter, r *http.Request) {
	targetID := s.todayTargetID()
	resp := hintsResponse{Tier: guessCount(r) / guessesPerTier}

	if resp.Tier >= 1 {
		if p, err := s.api.Pokemon(targetID); err == nil {
			if p.Cries.Latest != "" {
				resp.Cry = s.cachedCry(targetID, p.Cries.Latest)
			}
			if resp.Tier >= 2 {
				for _, t := range p.Types {
					resp.Types = append(resp.Types, strings.ToUpper(t.Type.Name))
				}
			}
		}
	}

	if resp.Tier >= 3 {
		resp.Description = s.descriptions(targetID)
	}

	writeJSON(w, resp)
}

// cachedCry downloads the cry into static/ (once) and returns its public URL.
func (s *Server) cachedCry(id int, url string) string {
	name := fmt.Sprintf("hint_%d.ogg", id)
	path := filepath.Join(s.cfg.StaticDir, name)
	public := "/static/" + name

	if _, err := os.Stat(path); err == nil {
		return public
	}

	body, err := s.api.Download(url)
	if err != nil {
		return ""
	}
	defer body.Close()

	out, err := os.Create(path)
	if err != nil {
		return ""
	}
	defer out.Close()

	if _, err := io.Copy(out, body); err != nil {
		os.Remove(path)
		return ""
	}
	return public
}

// descriptions returns the first Pokédex entry found for each supported language.
func (s *Server) descriptions(id int) map[string]string {
	species, err := s.api.Species(id)
	if err != nil {
		return nil
	}

	allowed := make(map[string]bool, len(dataset.Languages))
	for _, l := range dataset.Languages {
		allowed[l] = true
	}

	out := make(map[string]string)
	for _, e := range species.FlavorTextEntries {
		lang := e.Language.Name
		if _, done := out[lang]; allowed[lang] && !done {
			out[lang] = cleanFlavorText(e.FlavorText)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cleanFlavorText(s string) string {
	s = strings.NewReplacer("\n", " ", "\f", " ").Replace(s)
	return strings.TrimSpace(s)
}
