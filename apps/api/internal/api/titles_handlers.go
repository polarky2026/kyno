package api

import (
	"context"
	"errors"
	"log/slog"

	"github.com/polarky2026/kyno/apps/api/internal/titles"
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
		return SearchTitles502JSONResponse{Error: "The catalog is unavailable right now. Try again in a moment."}, nil
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

// optionalString turns an empty string into nil, so the field is left out
// of the JSON instead of being sent as "".
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
