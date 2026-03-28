// Package scryfall provides a client for the Scryfall card search API
// with local file caching and automatic rate limiting.
package scryfall

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

// Card holds the Scryfall data needed for decklist rendering.
type Card struct {
	Name          string   `json:"name"`
	TypeLine      string   `json:"type_line"`
	ManaCost      string   `json:"mana_cost"`
	CMC           float64  `json:"cmc"`
	ColorIdentity []string `json:"color_identity"`
	Colors        []string `json:"colors"`
}

const (
	baseURL   = "https://api.scryfall.com"
	userAgent = "jumpforge/1.0 (MTG Jumpstart decklist formatter)"
	cacheTTL  = 7 * 24 * time.Hour // 1 week
)

// Client fetches card data from the Scryfall API with caching and rate limiting.
type Client struct {
	httpClient *http.Client
	cacheDir   string
	lastReq    time.Time
	log        zerolog.Logger
}

// NewClient creates a Scryfall client with default settings.
func NewClient(log zerolog.Logger) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cacheDir:   cacheDir(),
		log:        log.With().Str("component", "scryfall").Logger(),
	}
}

// FetchCard fetches a single card by exact name, using the cache when possible.
// See https://scryfall.com/docs/api/cards/named
func (c *Client) FetchCard(name string) (*Card, error) {
	requestURL := fmt.Sprintf("%s/cards/named?exact=%s", baseURL, url.QueryEscape(name))
	body, err := c.doRequest(requestURL, safeFileName(name))
	if err != nil {
		return nil, fmt.Errorf("fetch card %q: %w", name, err)
	}

	var card Card
	if err := json.Unmarshal(body, &card); err != nil {
		return nil, fmt.Errorf("parse card %q: %w", name, err)
	}
	return &card, nil
}
// doRequest performs an HTTP GET with caching, rate limiting, and retry.
// If cacheKey is non-empty, the response is cached and served from cache
// on subsequent calls.
func (c *Client) doRequest(requestURL string, cacheKey string) ([]byte, error) {
	if cacheKey != "" {
		if data, ok := readCacheFile(c.cacheDir, cacheKey, cacheTTL); ok {
			c.log.Debug().Str("cacheKey", cacheKey).Msg("cache hit")
			return data, nil
		}
	}

	body, err := c.executeWithRetry(requestURL)
	if err != nil {
		return nil, err
	}

	if cacheKey != "" {
		if err := writeCacheFile(c.cacheDir, cacheKey, body); err != nil {
			c.log.Warn().Err(err).Str("cacheKey", cacheKey).Msg("failed to write cache")
		}
	}

	return body, nil
}

// executeWithRetry sends a GET request with shared headers and rate limiting.
// Retries once on 429 after a 2s delay.
func (c *Client) executeWithRetry(requestURL string) ([]byte, error) {
	// Rate limit: 100ms between requests per Scryfall guidelines.
	if !c.lastReq.IsZero() {
		if wait := 100*time.Millisecond - time.Since(c.lastReq); wait > 0 {
			c.log.Debug().Dur("wait", wait).Msg("rate limit delay")
			time.Sleep(wait)
		}
	}

	body, status, err := c.sendRequest(requestURL)
	if err != nil {
		return nil, err
	}

	if status == http.StatusTooManyRequests {
		c.log.Warn().Str("url", requestURL).Msg("rate limited (429), retrying in 2s")
		time.Sleep(2 * time.Second)
		body, status, err = c.sendRequest(requestURL)
		if err != nil {
			return nil, err
		}
		if status == http.StatusTooManyRequests {
			return nil, fmt.Errorf("rate limited for %s after retry", requestURL)
		}
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", status, requestURL)
	}

	return body, nil
}

// sendRequest performs a single HTTP GET with standard headers.
func (c *Client) sendRequest(requestURL string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	c.lastReq = time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed for %s: %w", requestURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response for %s: %w", requestURL, err)
	}

	return body, resp.StatusCode, nil
}
