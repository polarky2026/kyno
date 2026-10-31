// Package titles is Kyno's catalog: movies and series with the data Kyno
// keeps from TMDB. It knows nothing about users, ratings or rankings.
package titles

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("titles: not found")
	ErrInvalidQuery      = errors.New("titles: invalid search query")
	ErrInvalidKind       = errors.New("titles: invalid kind")
	ErrInvalidTMDBID     = errors.New("titles: invalid TMDB id")
	ErrSourceUnavailable = errors.New("titles: catalog source unavailable")
)

type Kind string

const (
	KindMovie  Kind = "movie"
	KindSeries Kind = "series"
)

func (k Kind) Valid() bool {
	return k == KindMovie || k == KindSeries
}

// SearchResult is a title found in the external catalog. Search results are
// shown to the user but never stored.
type SearchResult struct {
	TMDBID        int64
	Kind          Kind
	Title         string
	OriginalTitle string
	Overview      string
	PosterPath    string // empty when TMDB has no poster
	Year          *int   // nil when the release date is unknown
}

// Season is a season of a series. Specials (season 0) are never part of the catalog.
type Season struct {
	Number       int
	Name         string
	AirDate      *time.Time // nil when TMDB doesn't know it yet
	EpisodeCount int
}

// AiredBy reports whether the season had started airing at the given time.
func (s Season) AiredBy(now time.Time) bool {
	return s.AirDate != nil && !s.AirDate.After(now)
}

// Title is a movie or series in Kyno's catalog.
type Title struct {
	ID            uuid.UUID
	Kind          Kind
	TMDBID        int64
	Title         string
	OriginalTitle string
	Overview      string
	PosterPath    string
	ReleaseDate   *time.Time // movies: release date; series: first air date
	Seasons       []Season   // ordered by number; always empty for movies
	SyncedAt      time.Time
}

// LibraryYear is the year the title is grouped under in the library: a
// movie's release year, or the year of a series' latest season that has
// already aired. A series never moves ahead to a season that hasn't aired
// yet. Nil means the year is unknown, and the title goes to "No year".
func (t Title) LibraryYear(now time.Time) *int {
	if t.Kind == KindMovie {
		return yearOfDate(t.ReleaseDate)
	}

	var latest *time.Time
	for _, season := range t.Seasons {
		if season.AiredBy(now) && (latest == nil || season.AirDate.After(*latest)) {
			latest = season.AirDate
		}
	}
	if latest == nil && t.ReleaseDate != nil && !t.ReleaseDate.After(now) {
		latest = t.ReleaseDate
	}
	return yearOfDate(latest)
}

// ExternalTitle is a title as the external catalog describes it, before it
// is stored in Kyno.
type ExternalTitle struct {
	Kind          Kind
	TMDBID        int64
	Title         string
	OriginalTitle string
	Overview      string
	PosterPath    string
	ReleaseDate   *time.Time
	Seasons       []Season
}

// Source is what the catalog needs from the external movie database.
type Source interface {
	// Search finds titles. A nil kind searches movies and series together.
	Search(ctx context.Context, query string, kind *Kind) ([]SearchResult, error)
	// GetTitle returns a title's details, or ErrNotFound when the external
	// catalog has no such title.
	GetTitle(ctx context.Context, kind Kind, tmdbID int64) (ExternalTitle, error)
}

// Store keeps Kyno's copy of the catalog.
type Store interface {
	// GetByID returns ErrNotFound when no title has this id.
	GetByID(ctx context.Context, id uuid.UUID) (Title, error)
	// GetByTMDBID returns ErrNotFound when the title isn't in the catalog yet.
	GetByTMDBID(ctx context.Context, kind Kind, tmdbID int64) (Title, error)
	// Save inserts or updates a title and its seasons in one transaction.
	// Seasons are added or updated but never deleted, so a season someone
	// has rated is never lost.
	Save(ctx context.Context, t ExternalTitle, syncedAt time.Time) (Title, error)
}

func yearOfDate(date *time.Time) *int {
	if date == nil {
		return nil
	}
	year := date.Year()
	return &year
}
