package titles

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polarky2026/kyno/apps/api/internal/db"
)

type PostgresStore struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

var _ Store = (*PostgresStore)(nil)

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool, queries: db.New(pool)}
}

func (s *PostgresStore) GetByID(ctx context.Context, id uuid.UUID) (Title, error) {
	row, err := s.queries.GetTitleByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Title{}, ErrNotFound
	}
	if err != nil {
		return Title{}, err
	}
	return s.withSeasons(ctx, s.queries, row)
}

func (s *PostgresStore) GetByTMDBID(ctx context.Context, kind Kind, tmdbID int64) (Title, error) {
	row, err := s.queries.GetTitleByTMDBID(ctx, db.GetTitleByTMDBIDParams{
		Kind:   string(kind),
		TmdbID: tmdbID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Title{}, ErrNotFound
	}
	if err != nil {
		return Title{}, err
	}
	return s.withSeasons(ctx, s.queries, row)
}

func (s *PostgresStore) Save(ctx context.Context, t ExternalTitle, syncedAt time.Time) (Title, error) {
	var saved Title

	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)

		row, err := q.UpsertTitle(ctx, db.UpsertTitleParams{
			Kind:          string(t.Kind),
			TmdbID:        t.TMDBID,
			Title:         t.Title,
			OriginalTitle: t.OriginalTitle,
			Overview:      t.Overview,
			PosterPath:    nullableString(t.PosterPath),
			ReleaseDate:   t.ReleaseDate,
			SyncedAt:      syncedAt,
		})
		if err != nil {
			return err
		}

		for _, season := range t.Seasons {
			err := q.UpsertSeason(ctx, db.UpsertSeasonParams{
				TitleID:      row.ID,
				SeasonNumber: int32(season.Number),
				Name:         season.Name,
				AirDate:      season.AirDate,
				EpisodeCount: int32(season.EpisodeCount),
			})
			if err != nil {
				return err
			}
		}

		saved, err = s.withSeasons(ctx, q, row)
		return err
	})

	return saved, err
}

func (s *PostgresStore) withSeasons(ctx context.Context, q *db.Queries, row db.Title) (Title, error) {
	rows, err := q.ListSeasons(ctx, row.ID)
	if err != nil {
		return Title{}, err
	}

	seasons := make([]Season, 0, len(rows))
	for _, r := range rows {
		seasons = append(seasons, Season{
			Number:       int(r.SeasonNumber),
			Name:         r.Name,
			AirDate:      r.AirDate,
			EpisodeCount: int(r.EpisodeCount),
		})
	}

	posterPath := ""
	if row.PosterPath != nil {
		posterPath = *row.PosterPath
	}

	return Title{
		ID:            row.ID,
		Kind:          Kind(row.Kind),
		TMDBID:        row.TmdbID,
		Title:         row.Title,
		OriginalTitle: row.OriginalTitle,
		Overview:      row.Overview,
		PosterPath:    posterPath,
		ReleaseDate:   row.ReleaseDate,
		Seasons:       seasons,
		SyncedAt:      row.SyncedAt,
	}, nil
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
