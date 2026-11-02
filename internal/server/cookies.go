package server

import (
	"net/http"
	"strconv"
	"time"
)

const (
	cookieSolved  = "solved"
	cookieGuesses = "guesses"
)

// guessCount returns the number of guesses already made today.
func guessCount(r *http.Request) int {
	c, err := r.Cookie(cookieGuesses)
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(c.Value)
	return n
}

func alreadySolved(r *http.Request) bool {
	c, err := r.Cookie(cookieSolved)
	return err == nil && c.Value == "true"
}

// cookieExpiry is next midnight (local time), or 5 minutes in dev mode.
func (s *Server) cookieExpiry() time.Time {
	now := time.Now()
	if s.cfg.DevMode {
		return now.Add(5 * time.Minute)
	}
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

func (s *Server) setCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  s.cookieExpiry(),
		HttpOnly: true,
	})
}
