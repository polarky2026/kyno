package titles

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

var testNow = time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)

// fakeSource is the external catalog, in memory.
type fakeSource struct {
	results     []SearchResult
	searchErr   error
	titles      map[int64]ExternalTitle
	getErr      error
	gotQuery    string
	gotKind     *Kind
	searchCalls int
	getCalls    int
}

func (f *fakeSource) Search(_ context.Context, query string, kind *Kind) ([]SearchResult, error) {
	f.searchCalls++
	f.gotQuery = query
	f.gotKind = kind
	return f.results, f.searchErr
}

func (f *fakeSource) GetTitle(_ context.Context, _ Kind, tmdbID int64) (ExternalTitle, error) {
	f.getCalls++
	if f.getErr != nil {
		return ExternalTitle{}, f.getErr
	}
	t, ok := f.titles[tmdbID]
	if !ok {
		return ExternalTitle{}, ErrNotFound
	}
	return t, nil
}

// fakeStore is the catalog, in memory. Like the real store, Save adds or
// updates seasons but never deletes them.
type fakeStore struct {
	titles    map[uuid.UUID]Title
	saveCalls int
}

func newFakeStore() *fakeStore {
	return &fakeStore{titles: map[uuid.UUID]Title{}}
}

func (f *fakeStore) GetByID(_ context.Context, id uuid.UUID) (Title, error) {
	t, ok := f.titles[id]
	if !ok {
		return Title{}, ErrNotFound
	}
	return t, nil
}

func (f *fakeStore) GetByTMDBID(_ context.Context, kind Kind, tmdbID int64) (Title, error) {
	for _, t := range f.titles {
		if t.Kind == kind && t.TMDBID == tmdbID {
			return t, nil
		}
	}
	return Title{}, ErrNotFound
}

func (f *fakeStore) Save(ctx context.Context, ext ExternalTitle, syncedAt time.Time) (Title, error) {
	f.saveCalls++

	t, err := f.GetByTMDBID(ctx, ext.Kind, ext.TMDBID)
	if errors.Is(err, ErrNotFound) {
		t = Title{ID: uuid.New()}
	}

	byNumber := map[int]Season{}
	for _, season := range t.Seasons {
		byNumber[season.Number] = season
	}
	for _, season := range ext.Seasons {
		byNumber[season.Number] = season
	}
	seasons := make([]Season, 0, len(byNumber))
	for n := 1; len(seasons) < len(byNumber); n++ {
		if season, ok := byNumber[n]; ok {
			seasons = append(seasons, season)
		}
	}

	t.Kind = ext.Kind
	t.TMDBID = ext.TMDBID
	t.Title = ext.Title
	t.OriginalTitle = ext.OriginalTitle
	t.Overview = ext.Overview
	t.PosterPath = ext.PosterPath
	t.ReleaseDate = ext.ReleaseDate
	t.Seasons = seasons
	t.SyncedAt = syncedAt
	f.titles[t.ID] = t
	return t, nil
}

func newTestService(source *fakeSource) (*Service, *fakeStore) {
	store := newFakeStore()
	svc := NewService(store, source)
	svc.now = func() time.Time { return testNow }
	return svc, store
}

func kindPtr(k Kind) *Kind { return &k }

func date(year int, month time.Month, day int) *time.Time {
	d := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &d
}

var dune = ExternalTitle{
	Kind: KindMovie, TMDBID: 438631, Title: "Dune", OriginalTitle: "Dune",
	PosterPath: "/dune.jpg", ReleaseDate: date(2021, 9, 15),
}

var thrones = ExternalTitle{
	Kind: KindSeries, TMDBID: 1399, Title: "Game of Thrones", OriginalTitle: "Game of Thrones",
	ReleaseDate: date(2011, 4, 17),
	Seasons: []Season{
		{Number: 0, Name: "Specials", AirDate: date(2010, 12, 5), EpisodeCount: 7},
		{Number: 1, Name: "Season 1", AirDate: date(2011, 4, 17), EpisodeCount: 10},
		{Number: 2, Name: "Season 2", AirDate: date(2012, 4, 1), EpisodeCount: 10},
	},
}

// Search

func TestSearchTrimsTheQueryAndPassesTheKind(t *testing.T) {
	source := &fakeSource{results: []SearchResult{{TMDBID: 438631, Kind: KindMovie, Title: "Dune"}}}
	svc, _ := newTestService(source)

	results, err := svc.Search(context.Background(), "  dune  ", kindPtr(KindMovie))
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if source.gotQuery != "dune" {
		t.Errorf("query = %q, want trimmed dune", source.gotQuery)
	}
	if source.gotKind == nil || *source.gotKind != KindMovie {
		t.Errorf("kind = %v, want movie", source.gotKind)
	}
	if len(results) != 1 || results[0].Title != "Dune" {
		t.Errorf("results = %+v", results)
	}
}

func TestSearchRejectsInvalidQueries(t *testing.T) {
	for _, query := range []string{"", "   ", strings.Repeat("a", 101)} {
		source := &fakeSource{}
		svc, _ := newTestService(source)

		if _, err := svc.Search(context.Background(), query, nil); !errors.Is(err, ErrInvalidQuery) {
			t.Errorf("Search(%d chars) error = %v; want ErrInvalidQuery", len(query), err)
		}
		if source.searchCalls != 0 {
			t.Error("an invalid query must not reach the source")
		}
	}
}

func TestSearchAcceptsQueriesUpToOneHundredCharacters(t *testing.T) {
	svc, _ := newTestService(&fakeSource{})

	if _, err := svc.Search(context.Background(), strings.Repeat("ñ", 100), nil); err != nil {
		t.Fatalf("a 100-character query must be valid: %v", err)
	}
}

func TestSearchRejectsUnknownKinds(t *testing.T) {
	source := &fakeSource{}
	svc, _ := newTestService(source)

	if _, err := svc.Search(context.Background(), "dune", kindPtr("person")); !errors.Is(err, ErrInvalidKind) {
		t.Fatalf("error = %v; want ErrInvalidKind", err)
	}
	if source.searchCalls != 0 {
		t.Error("an invalid kind must not reach the source")
	}
}

func TestSearchReportsSourceFailures(t *testing.T) {
	cause := errors.New("connection refused")
	svc, _ := newTestService(&fakeSource{searchErr: cause})

	_, err := svc.Search(context.Background(), "dune", nil)
	if !errors.Is(err, ErrSourceUnavailable) {
		t.Fatalf("error = %v; want ErrSourceUnavailable", err)
	}
	if !errors.Is(err, cause) {
		t.Error("the original cause must be kept for the logs")
	}
}

func TestSearchReturnsAtMostTwentyResults(t *testing.T) {
	svc, _ := newTestService(&fakeSource{results: make([]SearchResult, 30)})

	results, err := svc.Search(context.Background(), "dune", nil)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 20 {
		t.Fatalf("got %d results, want 20", len(results))
	}
}

// Import and Get

func TestImportStoresANewTitle(t *testing.T) {
	source := &fakeSource{titles: map[int64]ExternalTitle{1399: thrones}}
	svc, store := newTestService(source)

	title, err := svc.Import(context.Background(), KindSeries, 1399)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if title.ID == uuid.Nil || title.Title != "Game of Thrones" || !title.SyncedAt.Equal(testNow) {
		t.Errorf("title = %+v", title)
	}
	if len(title.Seasons) != 2 || title.Seasons[0].Number != 1 {
		t.Errorf("seasons = %+v; want seasons 1 and 2, without the specials", title.Seasons)
	}
	if store.saveCalls != 1 {
		t.Errorf("saved %d times, want 1", store.saveCalls)
	}
}

func TestImportTwiceReturnsTheSameTitle(t *testing.T) {
	source := &fakeSource{titles: map[int64]ExternalTitle{438631: dune}}
	svc, store := newTestService(source)

	first, _ := svc.Import(context.Background(), KindMovie, 438631)
	second, err := svc.Import(context.Background(), KindMovie, 438631)
	if err != nil {
		t.Fatalf("second Import: %v", err)
	}
	if first.ID != second.ID || len(store.titles) != 1 {
		t.Fatal("importing twice must not duplicate the title")
	}
	if source.getCalls != 1 {
		t.Errorf("source called %d times; a fresh title must not be fetched again", source.getCalls)
	}
}

func TestImportRefreshesAStaleTitle(t *testing.T) {
	source := &fakeSource{titles: map[int64]ExternalTitle{1399: thrones}}
	svc, _ := newTestService(source)
	_, _ = svc.Import(context.Background(), KindSeries, 1399)

	updated := thrones
	updated.Seasons = []Season{{Number: 3, Name: "Season 3", AirDate: date(2013, 3, 31), EpisodeCount: 10}}
	source.titles[1399] = updated
	svc.now = func() time.Time { return testNow.Add(25 * time.Hour) }

	title, err := svc.Import(context.Background(), KindSeries, 1399)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if len(title.Seasons) != 3 {
		t.Fatalf("seasons = %+v; want 1, 2 and the new 3", title.Seasons)
	}
	if !title.SyncedAt.Equal(testNow.Add(25 * time.Hour)) {
		t.Error("a refreshed title must record when it was synced")
	}
}

func TestImportServesTheStoredCopyWhenRefreshFails(t *testing.T) {
	source := &fakeSource{titles: map[int64]ExternalTitle{438631: dune}}
	svc, _ := newTestService(source)
	stored, _ := svc.Import(context.Background(), KindMovie, 438631)

	source.getErr = errors.New("TMDB is down")
	svc.now = func() time.Time { return testNow.Add(48 * time.Hour) }

	title, err := svc.Import(context.Background(), KindMovie, 438631)
	if err != nil {
		t.Fatalf("Import = %v; want the stored copy", err)
	}
	if title.ID != stored.ID || !title.SyncedAt.Equal(stored.SyncedAt) {
		t.Errorf("title = %+v; want the stored copy", title)
	}
}

func TestImportUnknownTitle(t *testing.T) {
	svc, store := newTestService(&fakeSource{titles: map[int64]ExternalTitle{}})

	if _, err := svc.Import(context.Background(), KindMovie, 42); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v; want ErrNotFound", err)
	}
	if store.saveCalls != 0 {
		t.Error("nothing must be stored")
	}
}

func TestImportReportsSourceFailuresForNewTitles(t *testing.T) {
	svc, _ := newTestService(&fakeSource{getErr: errors.New("TMDB is down")})

	if _, err := svc.Import(context.Background(), KindMovie, 42); !errors.Is(err, ErrSourceUnavailable) {
		t.Fatalf("error = %v; want ErrSourceUnavailable", err)
	}
}

func TestImportValidatesInput(t *testing.T) {
	source := &fakeSource{}
	svc, _ := newTestService(source)

	if _, err := svc.Import(context.Background(), "person", 42); !errors.Is(err, ErrInvalidKind) {
		t.Errorf("error = %v; want ErrInvalidKind", err)
	}
	if _, err := svc.Import(context.Background(), KindMovie, 0); !errors.Is(err, ErrInvalidTMDBID) {
		t.Errorf("error = %v; want ErrInvalidTMDBID", err)
	}
	if source.getCalls != 0 {
		t.Error("invalid input must not reach the source")
	}
}

func TestGetReturnsAStoredTitle(t *testing.T) {
	source := &fakeSource{titles: map[int64]ExternalTitle{438631: dune}}
	svc, _ := newTestService(source)
	imported, _ := svc.Import(context.Background(), KindMovie, 438631)

	title, err := svc.Get(context.Background(), imported.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if title.Title != "Dune" || source.getCalls != 1 {
		t.Errorf("title = %+v, source calls = %d", title, source.getCalls)
	}
}

func TestGetUnknownTitle(t *testing.T) {
	svc, _ := newTestService(&fakeSource{})

	if _, err := svc.Get(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v; want ErrNotFound", err)
	}
}

func TestGetRefreshesAStaleTitle(t *testing.T) {
	source := &fakeSource{titles: map[int64]ExternalTitle{438631: dune}}
	svc, _ := newTestService(source)
	imported, _ := svc.Import(context.Background(), KindMovie, 438631)

	renamed := dune
	renamed.Title = "Dune: Part One"
	source.titles[438631] = renamed
	svc.now = func() time.Time { return testNow.Add(25 * time.Hour) }

	title, err := svc.Get(context.Background(), imported.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if title.Title != "Dune: Part One" {
		t.Errorf("title = %q; want the refreshed name", title.Title)
	}
}
