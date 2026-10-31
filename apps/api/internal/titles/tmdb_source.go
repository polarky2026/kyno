package titles

import (
	"context"
	"errors"
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
				Year:          yearOfDate(dateOf(r.ReleaseDate)),
			})
		case "tv":
			results = append(results, SearchResult{
				TMDBID:        r.ID,
				Kind:          KindSeries,
				Title:         r.Name,
				OriginalTitle: r.OriginalName,
				Overview:      r.Overview,
				PosterPath:    r.PosterPath,
				Year:          yearOfDate(dateOf(r.FirstAirDate)),
			})
		}
		// People and any other media types are not part of the catalog.
	}
	return results, nil
}

func (s *TMDBSource) GetTitle(ctx context.Context, kind Kind, tmdbID int64) (ExternalTitle, error) {
	if kind == KindMovie {
		movie, err := s.client.GetMovie(ctx, tmdbID)
		if err != nil {
			return ExternalTitle{}, notFoundOr(err)
		}
		return ExternalTitle{
			Kind:          KindMovie,
			TMDBID:        movie.ID,
			Title:         movie.Title,
			OriginalTitle: movie.OriginalTitle,
			Overview:      movie.Overview,
			PosterPath:    movie.PosterPath,
			ReleaseDate:   dateOf(movie.ReleaseDate),
		}, nil
	}

	show, err := s.client.GetTV(ctx, tmdbID)
	if err != nil {
		return ExternalTitle{}, notFoundOr(err)
	}

	seasons := make([]Season, 0, len(show.Seasons))
	for _, season := range show.Seasons {
		seasons = append(seasons, Season{
			Number:       season.SeasonNumber,
			Name:         season.Name,
			AirDate:      dateOf(season.AirDate),
			EpisodeCount: season.EpisodeCount,
		})
	}

	return ExternalTitle{
		Kind:          KindSeries,
		TMDBID:        show.ID,
		Title:         show.Name,
		OriginalTitle: show.OriginalName,
		Overview:      show.Overview,
		PosterPath:    show.PosterPath,
		ReleaseDate:   dateOf(show.FirstAirDate),
		Seasons:       seasons,
	}, nil
}

func notFoundOr(err error) error {
	if errors.Is(err, tmdb.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

// dateOf parses a TMDB date ("2006-01-02"). It returns nil when TMDB doesn't
// know the date yet and sends an empty string.
func dateOf(date string) *time.Time {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return nil
	}
	return &t
}
