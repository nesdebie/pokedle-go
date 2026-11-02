package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pokedle/internal/game"
)

// fakePokeAPI serves /pokemon/{id}, /pokemon-species/{id} and a cry file.
func fakePokeAPI(t *testing.T) *httptest.Server {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var id int
		switch {
		case r.URL.Path == "/cry.ogg":
			w.Write([]byte("OggS"))
		case sscan(r.URL.Path, "/pokemon/%d", &id):
			fmt.Fprintf(w, `{"id":%d,"name":"poke-%d","height":%d,"weight":%d,
				"types":[{"slot":1,"type":{"name":"fire"}}],
				"sprites":{"front_default":"front.png","other":{"official-artwork":{"front_default":"art.png"}}},
				"cries":{"latest":"%s/cry.ogg"}}`, id, id, id, id*10, ts.URL)
		case sscan(r.URL.Path, "/pokemon-species/%d", &id):
			w.Write([]byte(`{"flavor_text_entries":[
				{"flavor_text":"Hello\nworld","language":{"name":"en"}},
				{"flavor_text":"Second","language":{"name":"en"}},
				{"flavor_text":"Bonjour","language":{"name":"fr"}},
				{"flavor_text":"Konnichiwa","language":{"name":"ja"}}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(ts.Close)
	return ts
}

func sscan(path, format string, id *int) bool {
	n, _ := fmt.Sscanf(path, format, id)
	return n == 1
}

func newTestServer(t *testing.T) *Server {
	t.Chdir(filepath.Join("..", "..")) // CSV paths are relative to the project root
	static := t.TempDir()
	os.WriteFile(filepath.Join(static, "index.html"), []byte("<html>"), 0o644)

	srv, err := New(Config{StaticDir: static, EnvFile: "does-not-exist", PokeAPI: fakePokeAPI(t).URL})
	if err != nil {
		t.Fatal(err)
	}
	srv.secret = "test"
	return srv
}

func do(t *testing.T, h http.Handler, method, path, body string, cookies ...*http.Cookie) (*httptest.ResponseRecorder, map[string]any) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec, out
}

func TestGuessFlow(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	_, out := do(t, h, "POST", "/api/guess", `{"guess":"not-a-pokemon"}`)
	if out["ok"] != false || out["error"] != "Incorrect Pokémon name" {
		t.Fatalf("unknown name: %v", out)
	}

	// Guess the right Pokémon (upper-cased to check case insensitivity).
	targetID := srv.todayTargetID()
	name := strings.ToUpper(srv.names.EnglishName(targetID))
	rec, out := do(t, h, "POST", "/api/guess", fmt.Sprintf(`{"guess":%q}`, name),
		&http.Cookie{Name: cookieGuesses, Value: "2"})
	if out["ok"] != true || out["correct"] != true || out["guessCounter"] != 3.0 {
		t.Fatalf("correct guess: %v", out)
	}
	guess := out["guess"].(map[string]any)
	if guess["name"] != srv.names.EnglishName(targetID) || guess["sprite"] != "art.png" {
		t.Errorf("guess card: %v", guess)
	}
	if _, ok := guess["position"]; !ok {
		t.Error("guess card must contain position")
	}
	hints := out["hints"].(map[string]any)
	if hints["weightHint"] != fmt.Sprintf("%.1fkg", float64(targetID)) || hints["type2"] != "(none)" {
		t.Errorf("hints: %v", hints)
	}
	if out["reveal"].(map[string]any)["id"] != float64(targetID) {
		t.Errorf("reveal: %v", out["reveal"])
	}

	solved := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == cookieSolved && c.Value == "true" && c.Expires.After(time.Now()) {
			solved = true
		}
	}
	if !solved {
		t.Fatal("solved cookie not set")
	}

	_, out = do(t, h, "POST", "/api/guess", `{"guess":"pikachu"}`, &http.Cookie{Name: cookieSolved, Value: "true"})
	if out["correct"] != true || out["ok"] != false {
		t.Errorf("already solved: %v", out)
	}
}

func TestHintsTiers(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	_, out := do(t, h, "GET", "/api/hints", "")
	if out["tier"] != 0.0 || out["cry"] != nil || out["types"] != nil {
		t.Errorf("tier 0: %v", out)
	}

	_, out = do(t, h, "GET", "/api/hints", "", &http.Cookie{Name: cookieGuesses, Value: "9"})
	if out["tier"] != 3.0 {
		t.Fatalf("tier 3: %v", out)
	}
	cry := fmt.Sprintf("/static/hint_%d.ogg", srv.todayTargetID())
	if out["cry"] != cry {
		t.Errorf("cry = %v, want %s", out["cry"], cry)
	}
	if _, err := os.Stat(filepath.Join(srv.cfg.StaticDir, filepath.Base(cry))); err != nil {
		t.Errorf("cry not downloaded: %v", err)
	}
	if types := out["types"].([]any); len(types) != 1 || types[0] != "FIRE" {
		t.Errorf("types: %v", out["types"])
	}
	desc := out["description"].(map[string]any)
	if desc["en"] != "Hello world" || desc["fr"] != "Bonjour" || desc["ja"] != nil {
		t.Errorf("description: %v", desc)
	}
}

func TestTodayAndSuggest(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	_, out := do(t, h, "GET", "/api/today", "")
	if out["date"] != game.DayKey(time.Now()) || out["max"] != float64(srv.names.Len()) {
		t.Errorf("today: %v", out)
	}

	rec, _ := do(t, h, "GET", "/api/suggest", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("suggest GET status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/api/suggest", strings.NewReader(`{"query":"bulb"}`)))
	if !strings.Contains(rec.Body.String(), "Bulbizarre") {
		t.Errorf("suggest: %s", rec.Body.String())
	}

	rec, _ = do(t, h, "GET", "/", "")
	if rec.Code != http.StatusOK || rec.Body.String() != "<html>" {
		t.Errorf("index: %d %q", rec.Code, rec.Body.String())
	}
}
