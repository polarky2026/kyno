package titles

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeSource struct {
	results   []SearchResult
	err       error
	gotQuery  string
	gotKind   *Kind
	callCount int
}

func (f *fakeSource) Search(_ context.Context, query string, kind *Kind) ([]SearchResult, error) {
	f.callCount++
	f.gotQuery = query
	f.gotKind = kind
	return f.results, f.err
}

func kindPtr(k Kind) *Kind { return &k }

func TestSearchTrimsTheQueryAndPassesTheKind(t *testing.T) {
	source := &fakeSource{results: []SearchResult{{TMDBID: 438631, Kind: KindMovie, Title: "Dune"}}}
	svc := NewService(source)

	results, err := svc.Search(context.Background(), "  dune  ", kindPtr(KindMovie))
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if source.gotQuery != "dune" {
		t.Errorf("query = %q, want trimmed dune", source.gotQuery)
	}
	if source.gotKind == nil || *source.gotKind != KindMovie {
		t.Errorf("kind = %v, want movie", source.gotKind)
	}
	if len(results) != 1 || results[0].Title != "Dune" {
		t.Errorf("results = %+v", results)
	}
}

func TestSearchRejectsInvalidQueries(t *testing.T) {
	for _, query := range []string{"", "   ", strings.Repeat("a", 101)} {
		source := &fakeSource{}
		svc := NewService(source)

		if _, err := svc.Search(context.Background(), query, nil); !errors.Is(err, ErrInvalidQuery) {
			t.Errorf("Search(%d chars) error = %v; want ErrInvalidQuery", len(query), err)
		}
		if source.callCount != 0 {
			t.Error("an invalid query must not reach the source")
		}
	}
}

func TestSearchAcceptsQueriesUpToOneHundredCharacters(t *testing.T) {
	svc := NewService(&fakeSource{})

	if _, err := svc.Search(context.Background(), strings.Repeat("ñ", 100), nil); err != nil {
		t.Fatalf("a 100-character query must be valid: %v", err)
	}
}

func TestSearchRejectsUnknownKinds(t *testing.T) {
	source := &fakeSource{}
	svc := NewService(source)

	if _, err := svc.Search(context.Background(), "dune", kindPtr("person")); !errors.Is(err, ErrInvalidKind) {
		t.Fatalf("error = %v; want ErrInvalidKind", err)
	}
	if source.callCount != 0 {
		t.Error("an invalid kind must not reach the source")
	}
}

func TestSearchReportsSourceFailures(t *testing.T) {
	cause := errors.New("connection refused")
	svc := NewService(&fakeSource{err: cause})

	_, err := svc.Search(context.Background(), "dune", nil)
	if !errors.Is(err, ErrSourceUnavailable) {
		t.Fatalf("error = %v; want ErrSourceUnavailable", err)
	}
	if !errors.Is(err, cause) {
		t.Error("the original cause must be kept for the logs")
	}
}

func TestSearchReturnsAtMostTwentyResults(t *testing.T) {
	many := make([]SearchResult, 30)
	svc := NewService(&fakeSource{results: many})

	results, err := svc.Search(context.Background(), "dune", nil)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 20 {
		t.Fatalf("got %d results, want 20", len(results))
	}
}
