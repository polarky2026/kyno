// Package titles is Kyno's catalog: movies and series with the data Kyno
// keeps from TMDB. It knows nothing about users, ratings or rankings.
package titles

import (
	"context"
	"errors"
)

var (
	ErrInvalidQuery      = errors.New("titles: invalid search query")
	ErrInvalidKind       = errors.New("titles: invalid kind")
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

// Source is what the catalog needs from the external movie database.
type Source interface {
	// Search finds titles. A nil kind searches movies and series together.
	Search(ctx context.Context, query string, kind *Kind) ([]SearchResult, error)
}
