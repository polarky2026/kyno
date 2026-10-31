package titles

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxQueryLength   = 100
	maxSearchResults = 20
)

type Service struct {
	source Source
}

func NewService(source Source) *Service {
	return &Service{source: source}
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
