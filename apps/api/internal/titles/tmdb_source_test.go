package titles

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/polarky2026/kyno/apps/api/internal/platform/tmdb"
)

func TestTMDBSourceMapsMoviesAndShowsAndSkipsPeople(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search/multi" {
			t.Errorf("path = %q, want /search/multi", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"results":[
			{"id":438631,"media_type":"movie","title":"Dune","original_title":"Dune","overview":"Paul…","poster_path":"/dune.jpg","release_date":"2021-09-15"},
			{"id":1399,"media_type":"tv","name":"Game of Thrones","original_name":"Game of Thrones","first_air_date":"2011-04-17"},
			{"id":7,"media_type":"person","name":"Denis Villeneuve"},
			{"id":999,"media_type":"movie","title":"Untitled Sequel","release_date":""}
		]}`))
	}))
	defer server.Close()

	source := NewTMDBSource(tmdb.NewClient("key", tmdb.WithBaseURL(server.URL)))
	results, err := source.Search(context.Background(), "dune", nil)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("got %d results, want 3 (the person is skipped): %+v", len(results), results)
	}

	movie := results[0]
	if movie.Kind != KindMovie || movie.TMDBID != 438631 || movie.Title != "Dune" ||
		movie.PosterPath != "/dune.jpg" || movie.Year == nil || *movie.Year != 2021 {
		t.Errorf("movie = %+v", movie)
	}

	show := results[1]
	if show.Kind != KindSeries || show.Title != "Game of Thrones" || show.Year == nil || *show.Year != 2011 {
		t.Errorf("show = %+v", show)
	}

	if results[2].Year != nil {
		t.Errorf("a title without a date must have no year, got %d", *results[2].Year)
	}
}

func TestTMDBSourceGetTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/movie/438631":
			_, _ = w.Write([]byte(`{"id":438631,"title":"Dune","original_title":"Dune","release_date":"2021-09-15","poster_path":"/dune.jpg"}`))
		case "/tv/1399":
			_, _ = w.Write([]byte(`{"id":1399,"name":"Game of Thrones","original_name":"Game of Thrones","first_air_date":"2011-04-17","seasons":[
				{"season_number":0,"name":"Specials","air_date":"2010-12-05","episode_count":7},
				{"season_number":1,"name":"Season 1","air_date":"2011-04-17","episode_count":10},
				{"season_number":2,"name":"Season 2","air_date":"","episode_count":0}
			]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	source := NewTMDBSource(tmdb.NewClient("key", tmdb.WithBaseURL(server.URL)))

	movie, err := source.GetTitle(context.Background(), KindMovie, 438631)
	if err != nil {
		t.Fatalf("GetTitle(movie): %v", err)
	}
	if movie.Kind != KindMovie || movie.Title != "Dune" || movie.ReleaseDate == nil ||
		movie.ReleaseDate.Year() != 2021 || len(movie.Seasons) != 0 {
		t.Errorf("movie = %+v", movie)
	}

	show, err := source.GetTitle(context.Background(), KindSeries, 1399)
	if err != nil {
		t.Fatalf("GetTitle(series): %v", err)
	}
	if show.Kind != KindSeries || show.Title != "Game of Thrones" || len(show.Seasons) != 3 {
		t.Fatalf("show = %+v", show)
	}
	if show.Seasons[2].AirDate != nil {
		t.Error("a season without an air date must have a nil AirDate")
	}

	if _, err := source.GetTitle(context.Background(), KindMovie, 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown title error = %v; want ErrNotFound", err)
	}
}

func TestTMDBSourceUsesTheEndpointForTheKind(t *testing.T) {
	cases := map[Kind]string{KindMovie: "/search/movie", KindSeries: "/search/tv"}

	for kind, wantPath := range cases {
		var gotPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			_, _ = w.Write([]byte(`{"results":[{"id":1,"title":"A","name":"A"}]}`))
		}))

		source := NewTMDBSource(tmdb.NewClient("key", tmdb.WithBaseURL(server.URL)))
		results, err := source.Search(context.Background(), "a", &kind)
		server.Close()

		if err != nil {
			t.Fatalf("Search(%s): %v", kind, err)
		}
		if gotPath != wantPath {
			t.Errorf("Search(%s) path = %q, want %q", kind, gotPath, wantPath)
		}
		if len(results) != 1 || results[0].Kind != kind {
			t.Errorf("Search(%s) results = %+v", kind, results)
		}
	}
}
