package tmdb

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testAPIKey = "secret-test-key"

func TestSearchMultiSendsTheExpectedRequest(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = map[string]string{}
		for key := range r.URL.Query() {
			gotQuery[key] = r.URL.Query().Get(key)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[
			{"id":438631,"media_type":"movie","title":"Dune","original_title":"Dune","release_date":"2021-09-15","poster_path":"/d5NXSklXo0qyIYkgV94XAgMIckC.jpg"},
			{"id":1,"media_type":"person","name":"Someone"}
		]}`))
	}))
	defer server.Close()

	client := NewClient(testAPIKey, WithBaseURL(server.URL))
	results, err := client.SearchMulti(context.Background(), "dune")
	if err != nil {
		t.Fatalf("SearchMulti: %v", err)
	}

	if gotPath != "/search/multi" {
		t.Errorf("path = %q, want /search/multi", gotPath)
	}
	want := map[string]string{
		"query":         "dune",
		"api_key":       testAPIKey,
		"language":      "en-US",
		"include_adult": "false",
		"page":          "1",
	}
	for key, value := range want {
		if gotQuery[key] != value {
			t.Errorf("query param %s = %q, want %q", key, gotQuery[key], value)
		}
	}

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if r := results[0]; r.ID != 438631 || r.MediaType != "movie" || r.Title != "Dune" || r.ReleaseDate != "2021-09-15" {
		t.Errorf("first result = %+v", r)
	}
}

func TestSingleTypeSearchesFillInTheMediaType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":1399,"name":"Game of Thrones","poster_path":null}]}`))
	}))
	defer server.Close()

	client := NewClient(testAPIKey, WithBaseURL(server.URL))

	shows, err := client.SearchTV(context.Background(), "thrones")
	if err != nil {
		t.Fatalf("SearchTV: %v", err)
	}
	if len(shows) != 1 || shows[0].MediaType != "tv" || shows[0].Name != "Game of Thrones" || shows[0].PosterPath != "" {
		t.Errorf("shows = %+v", shows)
	}

	movies, err := client.SearchMovies(context.Background(), "thrones")
	if err != nil {
		t.Fatalf("SearchMovies: %v", err)
	}
	if len(movies) != 1 || movies[0].MediaType != "movie" {
		t.Errorf("movies = %+v", movies)
	}
}

func TestGetMovieAndGetTV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("api_key") != testAPIKey || r.URL.Query().Get("language") != "en-US" {
			t.Errorf("missing api_key or language in %s", r.URL)
		}
		switch r.URL.Path {
		case "/movie/438631":
			_, _ = w.Write([]byte(`{"id":438631,"title":"Dune","original_title":"Dune","release_date":"2021-09-15","poster_path":"/dune.jpg"}`))
		case "/tv/1399":
			_, _ = w.Write([]byte(`{"id":1399,"name":"Game of Thrones","first_air_date":"2011-04-17","seasons":[
				{"season_number":0,"name":"Specials","air_date":"2010-12-05","episode_count":7},
				{"season_number":1,"name":"Season 1","air_date":"2011-04-17","episode_count":10}
			]}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(testAPIKey, WithBaseURL(server.URL))

	movie, err := client.GetMovie(context.Background(), 438631)
	if err != nil {
		t.Fatalf("GetMovie: %v", err)
	}
	if movie.Title != "Dune" || movie.ReleaseDate != "2021-09-15" || movie.PosterPath != "/dune.jpg" {
		t.Errorf("movie = %+v", movie)
	}

	show, err := client.GetTV(context.Background(), 1399)
	if err != nil {
		t.Fatalf("GetTV: %v", err)
	}
	if show.Name != "Game of Thrones" || len(show.Seasons) != 2 || show.Seasons[1].EpisodeCount != 10 {
		t.Errorf("show = %+v", show)
	}
}

func TestUnexpectedStatusIsAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := NewClient(testAPIKey, WithBaseURL(server.URL)).SearchMulti(context.Background(), "dune")

	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("error = %v; want a StatusError with 401", err)
	}
}

func TestNotFoundStatusIsErrNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := NewClient(testAPIKey, WithBaseURL(server.URL)).SearchMulti(context.Background(), "dune")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v; want ErrNotFound", err)
	}
}

func TestConnectionErrorsDoNotLeakTheAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Close() // nothing is listening anymore, so the request fails to connect

	_, err := NewClient(testAPIKey, WithBaseURL(server.URL)).SearchMulti(context.Background(), "dune")
	if err == nil {
		t.Fatal("expected a connection error")
	}
	if strings.Contains(err.Error(), testAPIKey) {
		t.Fatalf("error leaks the API key: %v", err)
	}
}
