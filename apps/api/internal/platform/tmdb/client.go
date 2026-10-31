// Package tmdb is a small client for The Movie Database (TMDB) API v3.
// It returns TMDB's own shapes; mapping them to Kyno's catalog happens in
// the titles package.
package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL  = "https://api.themoviedb.org/3"
	defaultLanguage = "en-US"
	requestTimeout  = 10 * time.Second
	maxResponseSize = 5 << 20
)

// ErrNotFound is returned when TMDB has no resource at the requested path.
var ErrNotFound = errors.New("tmdb: not found")

// StatusError is returned when TMDB answers with an unexpected status code.
type StatusError struct {
	Path       string
	StatusCode int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("tmdb: GET %s returned status %d", e.Path, e.StatusCode)
}

type Client struct {
	apiKey   string
	baseURL  string
	language string
	http     *http.Client
}

type Option func(*Client)

// WithBaseURL points the client at another server, such as a test server.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = baseURL }
}

func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:   apiKey,
		baseURL:  defaultBaseURL,
		language: defaultLanguage,
		http:     &http.Client{Timeout: requestTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SearchResult is one item from a TMDB search. Movies fill Title,
// OriginalTitle and ReleaseDate; TV shows fill Name, OriginalName and
// FirstAirDate. MediaType is "movie", "tv" or "person".
type SearchResult struct {
	ID            int64  `json:"id"`
	MediaType     string `json:"media_type"`
	Title         string `json:"title"`
	OriginalTitle string `json:"original_title"`
	Name          string `json:"name"`
	OriginalName  string `json:"original_name"`
	Overview      string `json:"overview"`
	PosterPath    string `json:"poster_path"`
	ReleaseDate   string `json:"release_date"`
	FirstAirDate  string `json:"first_air_date"`
}

type searchResponse struct {
	Results []SearchResult `json:"results"`
}

// SearchMulti searches movies, TV shows and people at once.
func (c *Client) SearchMulti(ctx context.Context, query string) ([]SearchResult, error) {
	return c.search(ctx, "/search/multi", query, "")
}

// SearchMovies searches only movies.
func (c *Client) SearchMovies(ctx context.Context, query string) ([]SearchResult, error) {
	return c.search(ctx, "/search/movie", query, "movie")
}

// SearchTV searches only TV shows.
func (c *Client) SearchTV(ctx context.Context, query string) ([]SearchResult, error) {
	return c.search(ctx, "/search/tv", query, "tv")
}

// search runs a search. Single-type endpoints don't return media_type, so
// mediaType fills it in for them.
func (c *Client) search(ctx context.Context, path, query, mediaType string) ([]SearchResult, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("include_adult", "false")
	params.Set("page", "1")

	var body searchResponse
	if err := c.get(ctx, path, params, &body); err != nil {
		return nil, err
	}

	if mediaType != "" {
		for i := range body.Results {
			body.Results[i].MediaType = mediaType
		}
	}
	return body.Results, nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, out any) error {
	params.Set("api_key", c.apiKey)
	params.Set("language", c.language)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("tmdb: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		// A *url.Error includes the full URL, and with it the API key.
		// Keep only the underlying cause so the key never reaches the logs.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return fmt.Errorf("tmdb: GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode != http.StatusOK:
		return &StatusError{Path: path, StatusCode: resp.StatusCode}
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseSize)).Decode(out); err != nil {
		return fmt.Errorf("tmdb: GET %s: decode response: %w", path, err)
	}
	return nil
}
