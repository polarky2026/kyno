package titles

import (
	"context"
	"time"

	"github.com/polarky2026/kyno/apps/api/internal/platform/tmdb"
)

// TMDBSource adapts the TMDB client to the catalog's Source interface.
type TMDBSource struct {
	client *tmdb.Client
}

var _ Source = (*TMDBSource)(nil)

func NewTMDBSource(client *tmdb.Client) *TMDBSource {
	return &TMDBSource{client: client}
}

func (s *TMDBSource) Search(ctx context.Context, query string, kind *Kind) ([]SearchResult, error) {
	var (
		raw []tmdb.SearchResult
		err error
	)
	switch {
	case kind == nil:
		raw, err = s.client.SearchMulti(ctx, query)
	case *kind == KindMovie:
		raw, err = s.client.SearchMovies(ctx, query)
	default:
		raw, err = s.client.SearchTV(ctx, query)
	}
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(raw))
	for _, r := range raw {
		switch r.MediaType {
		case "movie":
			results = append(results, SearchResult{
				TMDBID:        r.ID,
				Kind:          KindMovie,
				Title:         r.Title,
				OriginalTitle: r.OriginalTitle,
				Overview:      r.Overview,
				PosterPath:    r.PosterPath,
				Year:          yearOf(r.ReleaseDate),
			})
		case "tv":
			results = append(results, SearchResult{
				TMDBID:        r.ID,
				Kind:          KindSeries,
				Title:         r.Name,
				OriginalTitle: r.OriginalName,
				Overview:      r.Overview,
				PosterPath:    r.PosterPath,
				Year:          yearOf(r.FirstAirDate),
			})
		}
		// People and any other media types are not part of the catalog.
	}
	return results, nil
}

// yearOf returns the year of a TMDB date ("2006-01-02"), or nil when TMDB
// doesn't know the date yet.
func yearOf(date string) *int {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return nil
	}
	year := t.Year()
	return &year
}
