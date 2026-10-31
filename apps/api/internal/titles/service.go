package titles

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxQueryLength   = 100
	maxSearchResults = 20
	// refreshAfter is how long a stored title is trusted before it is
	// fetched again from the external catalog.
	refreshAfter = 24 * time.Hour
)

type Service struct {
	store  Store
	source Source
	now    func() time.Time
}

func NewService(store Store, source Source) *Service {
	return &Service{store: store, source: source, now: time.Now}
}

// Search finds movies and series in the external catalog. Nothing is stored.
func (s *Service) Search(ctx context.Context, query string, kind *Kind) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if length := utf8.RuneCountInString(query); length == 0 || length > maxQueryLength {
		return nil, ErrInvalidQuery
	}
	if kind != nil && !kind.Valid() {
		return nil, ErrInvalidKind
	}

	results, err := s.source.Search(ctx, query, kind)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSourceUnavailable, err)
	}

	if len(results) > maxSearchResults {
		results = results[:maxSearchResults]
	}
	return results, nil
}

// Import returns the catalog's copy of a title, bringing it in from the
// external catalog the first time and refreshing it when it is stale.
// Importing the same title twice never duplicates it.
func (s *Service) Import(ctx context.Context, kind Kind, tmdbID int64) (Title, error) {
	if !kind.Valid() {
		return Title{}, ErrInvalidKind
	}
	if tmdbID <= 0 {
		return Title{}, ErrInvalidTMDBID
	}

	existing, err := s.store.GetByTMDBID(ctx, kind, tmdbID)
	switch {
	case err == nil:
		return s.refreshIfStale(ctx, existing)
	case !errors.Is(err, ErrNotFound):
		return Title{}, fmt.Errorf("look up title: %w", err)
	}

	external, err := s.source.GetTitle(ctx, kind, tmdbID)
	if errors.Is(err, ErrNotFound) {
		return Title{}, ErrNotFound
	}
	if err != nil {
		return Title{}, fmt.Errorf("%w: %w", ErrSourceUnavailable, err)
	}

	return s.save(ctx, external)
}

// Get returns a title from the catalog, refreshing it when it is stale.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (Title, error) {
	title, err := s.store.GetByID(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return Title{}, ErrNotFound
	}
	if err != nil {
		return Title{}, fmt.Errorf("get title: %w", err)
	}
	return s.refreshIfStale(ctx, title)
}

// refreshIfStale fetches a stored title again when it is older than
// refreshAfter. If the external catalog fails, the stored copy is returned:
// slightly old catalog data is better than an error.
func (s *Service) refreshIfStale(ctx context.Context, title Title) (Title, error) {
	if s.now().Sub(title.SyncedAt) < refreshAfter {
		return title, nil
	}

	external, err := s.source.GetTitle(ctx, title.Kind, title.TMDBID)
	if err != nil {
		slog.Warn("could not refresh title; serving the stored copy",
			"title_id", title.ID, "tmdb_id", title.TMDBID, "error", err)
		return title, nil
	}

	return s.save(ctx, external)
}

func (s *Service) save(ctx context.Context, external ExternalTitle) (Title, error) {
	external.Seasons = withoutSpecials(external.Seasons)

	saved, err := s.store.Save(ctx, external, s.now())
	if err != nil {
		return Title{}, fmt.Errorf("save title: %w", err)
	}
	return saved, nil
}

// withoutSpecials drops season 0, which holds the specials. Specials are
// never rated and never count towards a series' progress.
func withoutSpecials(seasons []Season) []Season {
	kept := make([]Season, 0, len(seasons))
	for _, season := range seasons {
		if season.Number > 0 {
			kept = append(kept, season)
		}
	}
	return kept
}
