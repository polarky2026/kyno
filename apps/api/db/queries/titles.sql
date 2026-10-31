-- name: GetTitleByID :one
SELECT * FROM titles WHERE id = $1;

-- name: GetTitleByTMDBID :one
SELECT * FROM titles WHERE kind = $1 AND tmdb_id = $2;

-- name: UpsertTitle :one
INSERT INTO titles (kind, tmdb_id, title, original_title, overview, poster_path, release_date, synced_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (kind, tmdb_id) DO UPDATE
SET title          = EXCLUDED.title,
    original_title = EXCLUDED.original_title,
    overview       = EXCLUDED.overview,
    poster_path    = EXCLUDED.poster_path,
    release_date   = EXCLUDED.release_date,
    synced_at      = EXCLUDED.synced_at,
    updated_at     = now()
RETURNING *;

-- name: UpsertSeason :exec
INSERT INTO seasons (title_id, season_number, name, air_date, episode_count)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (title_id, season_number) DO UPDATE
SET name          = EXCLUDED.name,
    air_date      = EXCLUDED.air_date,
    episode_count = EXCLUDED.episode_count,
    updated_at    = now();

-- name: ListSeasons :many
SELECT * FROM seasons
WHERE title_id = $1
ORDER BY season_number;
