// Package pokeapi is a minimal client for https://pokeapi.co (v2),
// shared by the game server and the CSV generation scripts.
package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const BaseURL = "https://pokeapi.co/api/v2"

// Client wraps an http.Client with helpers for PokéAPI resources.
type Client struct {
	BaseURL string // defaults to the public PokéAPI; overridable for tests
	HTTP    *http.Client
}

// NewClient returns a client with a sane request timeout.
func NewClient() *Client {
	return &Client{BaseURL: BaseURL, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

// Get performs a GET request on an absolute URL or a path relative to BaseURL
// (e.g. "pokemon/25") and decodes the JSON body into target.
func (c *Client) Get(urlOrPath string, target any) error {
	resp, err := c.HTTP.Get(c.resolve(urlOrPath))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pokeapi: status %d for %s", resp.StatusCode, urlOrPath)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

// Exists reports whether the given resource returns HTTP 200.
func (c *Client) Exists(urlOrPath string) (bool, error) {
	resp, err := c.HTTP.Get(c.resolve(urlOrPath))
	if err != nil {
		return false, err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

// Download fetches a raw file (e.g. a cry .ogg) and returns its body.
// The caller must close it.
func (c *Client) Download(url string) (io.ReadCloser, error) {
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("pokeapi: status %d for %s", resp.StatusCode, url)
	}
	return resp.Body, nil
}

func (c *Client) resolve(urlOrPath string) string {
	if strings.HasPrefix(urlOrPath, "http://") || strings.HasPrefix(urlOrPath, "https://") {
		return urlOrPath
	}
	return strings.TrimSuffix(c.BaseURL, "/") + "/" + strings.TrimPrefix(urlOrPath, "/")
}

// Count returns the "count" field of a paginated resource list (e.g. "pokemon").
func (c *Client) Count(resource string) (int, error) {
	var list struct {
		Count int `json:"count"`
	}
	if err := c.Get(resource, &list); err != nil {
		return 0, err
	}
	return list.Count, nil
}

// FindMaxID binary-searches the highest valid ID of a resource in [low, high].
func (c *Client) FindMaxID(resource string, low, high int) int {
	maxValid := 0
	for low <= high {
		mid := (low + high) / 2
		ok, err := c.Exists(fmt.Sprintf("%s/%d", resource, mid))
		if err != nil {
			fmt.Printf("Error checking %s/%d: %v\n", resource, mid, err)
			return maxValid
		}
		if ok {
			maxValid = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return maxValid
}

// IDFromURL extracts the trailing numeric ID of a PokéAPI resource URL,
// e.g. "https://pokeapi.co/api/v2/pokemon-species/25/" -> 25.
func IDFromURL(url string) int {
	parts := strings.Split(strings.Trim(url, "/"), "/")
	id, _ := strconv.Atoi(parts[len(parts)-1])
	return id
}
