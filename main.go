
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
	"bufio"

	"golang.org/x/text/unicode/norm"
)

// ---------------- Data models from PokeAPI  ----------------

type Stat struct {
	BaseStat int `json:"base_stat"`
	StatInfo struct {
		Name string `json:"name"`
	} `json:"stat"`
}

type TypeEntry struct {
	Slot int `json:"slot"`
	Type struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"type"`
}

type Sprites struct {
	FrontDefault string `json:"front_default"`
	Other        map[string]struct {
		FrontDefault string `json:"front_default"`
		FrontShiny   string `json:"front_shiny"`
	} `json:"other"`
}

type Pokemon struct {
	ID      int         `json:"id"`
	Name    string      `json:"name"`
	Height  int         `json:"height"`
	Weight  int         `json:"weight"`
	Stats   []Stat      `json:"stats"`
	Types   []TypeEntry `json:"types"`
	Sprites Sprites     `json:"sprites"`
}

// ---------------- Local name mapping ----------------

type NamesRow struct {
	ID int
	EN string
	FR string
	DE string
	ES string
	IT string
}

type NameIndex struct {
	idByKey map[string]int // normalized name -> id
	rows    []NamesRow      // keep list for deterministic indexing
}

func removeAccents(s string) string {
	decomposed := norm.NFD.String(s)
	var result []rune
	for _, r := range decomposed {
		if !unicode.Is(unicode.Mn, r) {
			result = append(result, r)
		}
	}
	return string(result)
}

func normalizeKey(s string) string {
	return strings.ToLower(removeAccents(strings.TrimSpace(s)))
}

func loadNames(csvPath string) (*NameIndex, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("csv has no rows")
	}

	idx := &NameIndex{idByKey: make(map[string]int)}
	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) < 6 {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSpace(row[0]))
		if err != nil {
			continue
		}
		nr := NamesRow{
			ID: id,
			EN: row[1],
			FR: row[2],
			DE: row[3],
			ES: row[4],
			IT: row[5],
		}
		idx.rows = append(idx.rows, nr)
		for _, name := range row[1:6] {
			k := normalizeKey(name)
			if k != "" {
				idx.idByKey[k] = id
			}
		}
	}
	return idx, nil
}

func (n *NameIndex) maxIndex() int { return len(n.rows) }

func (n *NameIndex) idAt(i int) int {
	if i < 0 || i >= len(n.rows) {
		return 0
	}
	return n.rows[i].ID
}

// ---------------- Daily target logic ----------------

func dayKey(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}


func loadEnvKey(filename, key string) string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, key+"=") {
			return strings.TrimPrefix(line, key+"=")
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return ""
}

func pickDailyIndex(names *NameIndex, t time.Time) int {
	if names.maxIndex() == 0 {
		return 0
	}
	secret := loadEnvKey(".env", "POKEDLE_SECRET")
	msg := []byte(dayKey(t))
	var sum []byte
	if secret != "" {
		m := hmac.New(sha256.New, []byte(secret))
		m.Write(msg)
		sum = m.Sum(nil)
	} else {
		h := sha256.Sum256(msg)
		sum = h[:]
	}

	var v uint64 = 0
	for i := 0; i < 8; i++ {
		v = (v << 8) | uint64(sum[i])
	}
	return int(v % uint64(names.maxIndex()))
}

// ---------------- Server state ----------------

type Server struct {
	names    *NameIndex
	csvPath  string
	dataDir  string
	staticFS http.Handler
}

func must[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func NewServer() *Server {
	wd, _ := os.Getwd()
	dataDir := filepath.Join(wd, "data")
	csvPath := filepath.Join(dataDir, "pokemon_names_multilang.csv")

	names := must(loadNames(csvPath))
	staticFS := http.FileServer(http.Dir(filepath.Join(wd, "static")))

	return &Server{
		names:    names,
		csvPath:  csvPath,
		dataDir:  dataDir,
		staticFS: staticFS,
	}
}

// ---------------- Helpers ----------------

func fetchPokemon(id int) (*Pokemon, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%d", id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pokeapi status: %d", resp.StatusCode)
	}
	var p Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func typeSet(p *Pokemon) map[string]bool {
	m := make(map[string]bool)
	if p == nil {
		return m
	}
	for _, te := range p.Types {
		m[te.Type.Name] = true
	}
	return m
}

func intersectCount(a, b map[string]bool) int {
	n := 0
	for k := range a {
		if b[k] {
			n++
		}
	}
	return n
}

// ---------------- HTTP Handlers ----------------

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.staticFS.ServeHTTP(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join("static", "index.html"))
}

type GuessReq struct {
	Guess string `json:"guess"`
	Lang  string `json:"lang"`
}

type GuessResp struct {
	OK      bool              `json:"ok"`
	Error   string            `json:"error,omitempty"`
	Correct bool              `json:"correct"`
	Guess   map[string]any    `json:"guess"`
	Hints   map[string]any    `json:"hints"`
	Reveal  map[string]any    `json:"reveal,omitempty"`
}

func (s *Server) handleGuess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req GuessReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	key := normalizeKey(req.Guess)
	id, ok := s.names.idByKey[key]
	if !ok {
		writeJSON(w, GuessResp{OK: false, Error: "Nom de Pokémon introuvable", Correct: false})
		return
	}

	todayIdx := pickDailyIndex(s.names, time.Now().UTC())
	targetID := s.names.idAt(todayIdx)

	guessP, gErr := fetchPokemon(id)
	targetP, tErr := fetchPokemon(targetID)
	if gErr != nil || tErr != nil {
		writeJSON(w, GuessResp{OK: false, Error: "Erreur PokeAPI", Correct: false})
		return
	}

	typeMatch := intersectCount(typeSet(guessP), typeSet(targetP))
	idHint := 0
	if guessP.ID < targetP.ID {
		idHint = -1
	} else if guessP.ID > targetP.ID {
		idHint = 1
	}
	var weightHint string
	switch {
	case guessP.Weight < targetP.Weight:
		weightHint = "plus léger"
	case guessP.Weight > targetP.Weight:
		weightHint = "plus lourd"
	default:
		weightHint = "égal"
	}
	var heightHint string
	switch {
	case guessP.Height < targetP.Height:
		heightHint = "plus petit"
	case guessP.Height > targetP.Height:
		heightHint = "plus grand"
	default:
		heightHint = "égal"
	}

	sprite := guessP.Sprites.FrontDefault
	if oa, ok := guessP.Sprites.Other["official-artwork"]; ok {
		if oa.FrontDefault != "" {
			sprite = oa.FrontDefault
		}
	}

	resp := GuessResp{
		OK:      true,
		Correct: guessP.ID == targetP.ID,
		Guess: map[string]any{
			"id":     guessP.ID,
			"name":   guessP.Name,
			"types":  typeSet(guessP),
			"height": guessP.Height,
			"weight": guessP.Weight,
			"sprite": sprite,
		},
		Hints: map[string]any{
			"typeMatch":  typeMatch,
			"idHint":     idHint,
			"weightHint": weightHint,
			"heightHint": heightHint,
			"distance":   int(math.Abs(float64(targetP.ID - guessP.ID))),
		},
	}

	if resp.Correct {
		targetSprite := targetP.Sprites.FrontDefault
		if oa, ok := targetP.Sprites.Other["official-artwork"]; ok {
			if oa.FrontDefault != "" {
				targetSprite = oa.FrontDefault
			}
		}
		resp.Reveal = map[string]any{
			"id":     targetP.ID,
			"name":   targetP.Name,
			"types":  typeSet(targetP),
			"height": targetP.Height,
			"weight": targetP.Weight,
			"sprite": targetSprite,
		}
	}

	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func (s *Server) handleToday(w http.ResponseWriter, r *http.Request) {
	idx := pickDailyIndex(s.names, time.Now().UTC())
	writeJSON(w, map[string]any{
		"date":      dayKey(time.Now().UTC()),
		"index":     idx,
		"max":       s.names.maxIndex(),
		"remaining": s.names.maxIndex() - idx - 1,
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	srv := NewServer()

	http.HandleFunc("/", srv.handleIndex)
	http.Handle("/static/", http.StripPrefix("/static/", srv.staticFS))
	http.HandleFunc("/api/guess", srv.handleGuess)
	http.HandleFunc("/api/today", srv.handleToday)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Pokedle prototype running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
