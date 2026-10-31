package titles

import "testing"

func TestLibraryYear(t *testing.T) {
	cases := []struct {
		name  string
		title Title
		want  *int
	}{
		{
			name:  "movie uses its release year",
			title: Title{Kind: KindMovie, ReleaseDate: date(2021, 9, 15)},
			want:  intPtr(2021),
		},
		{
			name:  "movie without a release date has no year",
			title: Title{Kind: KindMovie},
			want:  nil,
		},
		{
			name: "series uses its latest aired season",
			title: Title{Kind: KindSeries, ReleaseDate: date(2011, 4, 17), Seasons: []Season{
				{Number: 1, AirDate: date(2011, 4, 17)},
				{Number: 2, AirDate: date(2019, 4, 14)},
			}},
			want: intPtr(2019),
		},
		{
			name: "series doesn't move ahead to a season that hasn't aired",
			title: Title{Kind: KindSeries, ReleaseDate: date(2022, 8, 21), Seasons: []Season{
				{Number: 1, AirDate: date(2022, 8, 21)},
				{Number: 2, AirDate: date(2024, 6, 16)},
				{Number: 3, AirDate: date(2027, 1, 10)},
			}},
			want: intPtr(2024),
		},
		{
			name: "series ignores seasons without a date",
			title: Title{Kind: KindSeries, Seasons: []Season{
				{Number: 1, AirDate: date(2020, 1, 1)},
				{Number: 2},
			}},
			want: intPtr(2020),
		},
		{
			name:  "series without season data uses its first air date",
			title: Title{Kind: KindSeries, ReleaseDate: date(2015, 3, 1)},
			want:  intPtr(2015),
		},
		{
			name:  "series that hasn't started airing has no year",
			title: Title{Kind: KindSeries, ReleaseDate: date(2027, 5, 1), Seasons: []Season{{Number: 1, AirDate: date(2027, 5, 1)}}},
			want:  nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.title.LibraryYear(testNow)
			switch {
			case tc.want == nil && got != nil:
				t.Errorf("LibraryYear = %d, want none", *got)
			case tc.want != nil && (got == nil || *got != *tc.want):
				t.Errorf("LibraryYear = %v, want %d", got, *tc.want)
			}
		})
	}
}

func intPtr(n int) *int { return &n }
