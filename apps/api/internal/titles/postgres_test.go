package titles

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// These tests run against a real, migrated database. They are skipped unless
// TEST_DATABASE_URL is set, for example:
//
//	TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/kyno go test ./internal/titles/...
//
// They only touch titles with TMDB ids from 900000000 on, and delete them
// before and after each test.
const testTMDBIDBase = 900_000_000

func newTestPostgresStore(t *testing.T) *PostgresStore {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	cleanup := func() {
		_, err := pool.Exec(context.Background(), "DELETE FROM titles WHERE tmdb_id >= $1", testTMDBIDBase)
		if err != nil {
			t.Fatalf("clean up test titles: %v", err)
		}
	}
	cleanup()
	t.Cleanup(func() {
		cleanup()
		pool.Close()
	})

	return NewPostgresStore(pool)
}

func TestPostgresStoreSavesAndReadsAMovie(t *testing.T) {
	store := newTestPostgresStore(t)
	ctx := context.Background()
	syncedAt := time.Now().UTC().Truncate(time.Microsecond)

	movie := ExternalTitle{
		Kind: KindMovie, TMDBID: testTMDBIDBase + 1, Title: "Test Movie", OriginalTitle: "Película de prueba",
		Overview: "A test.", ReleaseDate: date(2021, 9, 15),
	}

	saved, err := store.Save(ctx, movie, syncedAt)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID == uuid.Nil || saved.Title != "Test Movie" || saved.PosterPath != "" || !saved.SyncedAt.Equal(syncedAt) {
		t.Errorf("saved = %+v", saved)
	}
	if saved.ReleaseDate == nil || !saved.ReleaseDate.Equal(*movie.ReleaseDate) {
		t.Errorf("release date = %v, want %v", saved.ReleaseDate, movie.ReleaseDate)
	}

	byTMDB, err := store.GetByTMDBID(ctx, KindMovie, movie.TMDBID)
	if err != nil || byTMDB.ID != saved.ID {
		t.Fatalf("GetByTMDBID = %+v, %v", byTMDB, err)
	}
	byID, err := store.GetByID(ctx, saved.ID)
	if err != nil || byID.OriginalTitle != "Película de prueba" {
		t.Fatalf("GetByID = %+v, %v", byID, err)
	}

	if _, err := store.GetByTMDBID(ctx, KindSeries, movie.TMDBID); !errors.Is(err, ErrNotFound) {
		t.Errorf("the same TMDB id as a series must be a different title, got %v", err)
	}
	if _, err := store.GetByID(ctx, uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown id error = %v; want ErrNotFound", err)
	}
}

func TestPostgresStoreUpdatesSeasonsWithoutDeletingThem(t *testing.T) {
	store := newTestPostgresStore(t)
	ctx := context.Background()

	series := ExternalTitle{
		Kind: KindSeries, TMDBID: testTMDBIDBase + 2, Title: "Test Series", OriginalTitle: "Test Series",
		PosterPath: "/poster.jpg",
		Seasons: []Season{
			{Number: 1, Name: "Season 1", AirDate: date(2020, 1, 1), EpisodeCount: 8},
			{Number: 2, Name: "Season 2", EpisodeCount: 0},
		},
	}
	first, err := store.Save(ctx, series, time.Now())
	if err != nil {
		t.Fatalf("first Save: %v", err)
	}

	// TMDB now lists season 2 with a date and a new season 3, and no longer
	// lists season 1.
	series.Title = "Test Series (Renamed)"
	series.Seasons = []Season{
		{Number: 2, Name: "Season 2", AirDate: date(2021, 6, 1), EpisodeCount: 10},
		{Number: 3, Name: "Season 3", EpisodeCount: 0},
	}
	second, err := store.Save(ctx, series, time.Now())
	if err != nil {
		t.Fatalf("second Save: %v", err)
	}

	if second.ID != first.ID {
		t.Fatal("saving the same title again must update it, not create another")
	}
	if second.Title != "Test Series (Renamed)" || second.PosterPath != "/poster.jpg" {
		t.Errorf("title = %+v", second)
	}
	if len(second.Seasons) != 3 {
		t.Fatalf("seasons = %+v; want 1, 2 and 3 (season 1 must not be deleted)", second.Seasons)
	}
	if s := second.Seasons[1]; s.Number != 2 || s.EpisodeCount != 10 || s.AirDate == nil {
		t.Errorf("season 2 = %+v; want it updated", s)
	}
}
