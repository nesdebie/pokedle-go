/* ************************************************************************** */
/*                                                                            */
/*                                                        :::      ::::::::   */
/*   main.go                                            :+:      :+:    :+:   */
/*                                                    +:+ +:+         +:+     */
/*   By: nesdebie <nesdebie@student.s19.be>         +#+  +:+       +#+        */
/*                                                +#+#+#+#+#+   +#+           */
/*   Created: 2025/08/26 12:50:34 by nesdebie          #+#    #+#             */
/*   Updated: 2025/08/28 13:45:08 by nesdebie         ###   ########.fr       */
/*                                                                            */
/* ************************************************************************** */


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

var isDevMode bool

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

type EvolutionData struct {
    Position       int `json:"position"`
    IsFullyEvolved int `json:"is_fully_evolved"`
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

func loadEvolutionData(path string) (map[int]EvolutionData, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    rows, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    evoData := make(map[int]EvolutionData)

    for _, row := range rows[1:] { // skip header
        id, err1 := strconv.Atoi(row[0])
        pos, err2 := strconv.Atoi(row[1])
        evo, err3 := strconv.Atoi(row[2])
        if err1 != nil || err2 != nil || err3 != nil {
            continue
        }
        evoData[id] = EvolutionData{Position: pos, IsFullyEvolved: evo}
    }

    return evoData, nil
}


func loadGenerationMap(path string) (map[int]int, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    rows, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    genMap := make(map[int]int)

    for _, row := range rows[1:] {
        id, err1 := strconv.Atoi(row[0])
        gen, err2 := strconv.Atoi(row[1])
        if err1 != nil || err2 != nil {
            continue
        }
        genMap[id] = gen
    }

    return genMap, nil
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

func extractTypes(p *Pokemon) (string, string) {
    t1, t2 := "(none)", "(none)"
    for _, te := range p.Types {
        if te.Slot == 1 {
            t1 = strings.ToUpper(te.Type.Name)
        } else if te.Slot == 2 {
            t2 = strings.ToUpper(te.Type.Name)
        }
    }
    return t1, t2
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
	OK      		bool              `json:"ok"`
	Error   		string            `json:"error,omitempty"`
	Correct 		bool              `json:"correct"`
	Guess   		map[string]any    `json:"guess"`
	Hints   		map[string]any    `json:"hints"`
	Reveal  		map[string]any    `json:"reveal,omitempty"`
	GuessCounter	int				  `json:"guessCounter"`
}

func (s *Server) handleGuess(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("solved")
	if err == nil && cookie.Value == "true" {
		writeJSON(w, GuessResp{OK: false, Error: "You already found. Try tomorrow!", Correct: true})
		return
	}
	
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
		writeJSON(w, GuessResp{OK: false, Error: "Incorrect Pokémon name", Correct: false})
		return
	}

	cookie, err = r.Cookie("guesses")
	var guessCount int
	if err == nil {
		guessCount, _ = strconv.Atoi(cookie.Value)
	}
	guessCount++

	todayIdx := pickDailyIndex(s.names, time.Now().UTC())
	targetID := s.names.idAt(todayIdx)

	guessP, gErr := fetchPokemon(id)
	targetP, tErr := fetchPokemon(targetID)
	if gErr != nil || tErr != nil {
		writeJSON(w, GuessResp{OK: false, Error: "PokeAPI Error", Correct: false})
		return
	}

	evoMap, err := loadEvolutionData("data/pokemon_evolution_data.csv")
	if err != nil {
		panic(err)
	}
	
	guessEvo := evoMap[guessP.ID]
	targetEvo := evoMap[targetP.ID]
	

	guessType1, guessType2 := extractTypes(guessP)
	targetType1, targetType2 := extractTypes(targetP)

	genMap, err := loadGenerationMap("data/pokemon_id_gen.csv")
	if err != nil {
		panic(err)
	}
	
	guessGen := genMap[guessP.ID]
	targetGen := genMap[targetP.ID]
	
	weightHint := strconv.FormatFloat(float64(guessP.Weight)/10, 'f', 1, 64) + "kg"
	switch {
	case guessP.Weight < targetP.Weight:
		weightHint = ">" + weightHint
	case guessP.Weight > targetP.Weight:
		weightHint = "<" + weightHint
	}
	heightHint := strconv.Itoa(guessP.Height * 10) + "cm"
	switch {
	case guessP.Height < targetP.Height:
		heightHint = ">" + heightHint
	case guessP.Height > targetP.Height:
		heightHint = "<" + heightHint
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
			"types":  []string{guessType1, guessType2},
			"height": guessP.Height,
			"weight": guessP.Weight,
			"sprite": sprite,
			"position": guessEvo.Position,
			"isFullyEvolved": guessEvo.IsFullyEvolved,
		},
		Hints: map[string]any{
			"type1":      guessType1,
			"type2":      guessType2,
			"type1Match": (guessType1 == targetType1),
			"type2Match": (guessType2 == targetType2),
			"type1MatchWrongPlace": (guessType1 == targetType2),
			"type2MatchWrongPlace": (guessType2 == targetType1),
			"guessedGen":  guessGen,
			"correctGen":  targetGen,
			"weightHint": weightHint,
			"heightHint": heightHint,
			"guessPosition": guessEvo.Position,
			"targetPosition": targetEvo.Position,
			"guessFullyEvolved": guessEvo.IsFullyEvolved,
			"targetFullyEvolved": targetEvo.IsFullyEvolved,
			"distance":   int(math.Abs(float64(targetP.ID - guessP.ID))),
		},
		GuessCounter: guessCount,
	}
	

	now := time.Now()
	midnight := time.Date(
		now.Year(), now.Month(), now.Day()+1,
		0, 0, 0, 0,
		now.Location(),
	)
	if isDevMode {
		midnight = time.Now().Add(1 * time.Minute)
	}

	if resp.Correct {
		http.SetCookie(w, &http.Cookie{
			Name:     "solved",
			Value:    "true",
			Path:     "/",
			Expires:  midnight,
			HttpOnly: true,
		})
		targetSprite := targetP.Sprites.FrontDefault
		if oa, ok := targetP.Sprites.Other["official-artwork"]; ok {
			if oa.FrontDefault != "" {
				targetSprite = oa.FrontDefault
			}
		}
		resp.Reveal = map[string]any{
			"id":     targetP.ID,
			"name":   targetP.Name,
			"types":  []string{targetType1, targetType2},
			"height": targetP.Height,
			"weight": targetP.Weight,
			"sprite": targetSprite,
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "guesses",
		Value:    strconv.Itoa(guessCount),
		Path:     "/",
		Expires:  midnight,
		HttpOnly: true,
	})
	
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

	cookie, err := r.Cookie("guesses")
	var guessCount int
	if err == nil {
		guessCount, _ = strconv.Atoi(cookie.Value)
	}
	
	writeJSON(w, map[string]any{
		"date":      dayKey(time.Now().UTC()),
		"index":     idx,
		"max":       s.names.maxIndex(),
		"remaining": s.names.maxIndex() - idx - 1,
		"guessCounter" : guessCount,
	})
}

func main() {
	isDevMode = false
	if len(os.Args) == 2 && os.Args[1] == "dev" {
		isDevMode = true
	}

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
