package api

import (
	"context"
	"errors"
	"log/slog"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/polarky2026/kyno/apps/api/internal/titles"
)

const (
	titleNotFound      = "We couldn't find that title."
	catalogUnavailable = "The catalog is unavailable right now. Try again in a moment."
)

func (s *Server) SearchTitles(ctx context.Context, req SearchTitlesRequestObject) (SearchTitlesResponseObject, error) {
	if sessionFrom(ctx).result == nil {
		return SearchTitles401JSONResponse{Error: signInRequired}, nil
	}

	var kind *titles.Kind
	if req.Params.Kind != nil {
		k := titles.Kind(*req.Params.Kind)
		kind = &k
	}

	results, err := s.titles.Search(ctx, req.Params.Query, kind)
	switch {
	case errors.Is(err, titles.ErrInvalidQuery):
		return SearchTitles400JSONResponse{Error: "Search for 1 to 100 characters."}, nil
	case errors.Is(err, titles.ErrInvalidKind):
		return SearchTitles400JSONResponse{Error: "Kind must be movie or series."}, nil
	case errors.Is(err, titles.ErrSourceUnavailable):
		slog.Error("title search failed", "error", err)
		return SearchTitles502JSONResponse{Error: catalogUnavailable}, nil
	case err != nil:
		return nil, err
	}

	body := SearchTitles200JSONResponse{Results: make([]TitleSearchResult, 0, len(results))}
	for _, r := range results {
		body.Results = append(body.Results, TitleSearchResult{
			TmdbId:        r.TMDBID,
			Kind:          TitleKind(r.Kind),
			Title:         r.Title,
			OriginalTitle: r.OriginalTitle,
			Overview:      r.Overview,
			PosterPath:    optionalString(r.PosterPath),
			Year:          r.Year,
		})
	}
	return body, nil
}

func (s *Server) ImportTitle(ctx context.Context, req ImportTitleRequestObject) (ImportTitleResponseObject, error) {
	if sessionFrom(ctx).result == nil {
		return ImportTitle401JSONResponse{Error: signInRequired}, nil
	}
	if req.Body == nil {
		return ImportTitle400JSONResponse{Error: invalidBody}, nil
	}

	title, err := s.titles.Import(ctx, titles.Kind(req.Body.Kind), req.Body.TmdbId)
	switch {
	case errors.Is(err, titles.ErrInvalidKind):
		return ImportTitle400JSONResponse{Error: "Kind must be movie or series."}, nil
	case errors.Is(err, titles.ErrInvalidTMDBID):
		return ImportTitle400JSONResponse{Error: "tmdbId must be a positive number."}, nil
	case errors.Is(err, titles.ErrNotFound):
		return ImportTitle404JSONResponse{Error: titleNotFound}, nil
	case errors.Is(err, titles.ErrSourceUnavailable):
		slog.Error("title import failed", "error", err)
		return ImportTitle502JSONResponse{Error: catalogUnavailable}, nil
	case err != nil:
		return nil, err
	}

	return ImportTitle200JSONResponse(toAPITitle(title, time.Now())), nil
}

func (s *Server) GetTitle(ctx context.Context, req GetTitleRequestObject) (GetTitleResponseObject, error) {
	if sessionFrom(ctx).result == nil {
		return GetTitle401JSONResponse{Error: signInRequired}, nil
	}

	title, err := s.titles.Get(ctx, req.Id)
	if errors.Is(err, titles.ErrNotFound) {
		return GetTitle404JSONResponse{Error: titleNotFound}, nil
	}
	if err != nil {
		return nil, err
	}

	return GetTitle200JSONResponse(toAPITitle(title, time.Now())), nil
}

func toAPITitle(t titles.Title, now time.Time) Title {
	seasons := make([]Season, 0, len(t.Seasons))
	for _, season := range t.Seasons {
		seasons = append(seasons, Season{
			Number:       season.Number,
			Name:         season.Name,
			AirDate:      optionalDate(season.AirDate),
			EpisodeCount: season.EpisodeCount,
			Aired:        season.AiredBy(now),
		})
	}

	return Title{
		Id:            t.ID,
		Kind:          TitleKind(t.Kind),
		TmdbId:        t.TMDBID,
		Title:         t.Title,
		OriginalTitle: t.OriginalTitle,
		Overview:      t.Overview,
		PosterPath:    optionalString(t.PosterPath),
		ReleaseDate:   optionalDate(t.ReleaseDate),
		LibraryYear:   t.LibraryYear(now),
		Seasons:       seasons,
	}
}

func optionalDate(t *time.Time) *openapi_types.Date {
	if t == nil {
		return nil
	}
	return &openapi_types.Date{Time: *t}
}

// optionalString turns an empty string into nil, so the field is left out
// of the JSON instead of being sent as "".
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
