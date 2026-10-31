-- +goose Up
CREATE TABLE titles (
    id             uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    kind           text        NOT NULL CHECK (kind IN ('movie', 'series')),
    tmdb_id        bigint      NOT NULL CHECK (tmdb_id > 0),
    title          text        NOT NULL,
    original_title text        NOT NULL,
    overview       text        NOT NULL DEFAULT '',
    poster_path    text,
    release_date   date,
    synced_at      timestamptz NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (kind, tmdb_id)
);

CREATE TABLE seasons (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    title_id      uuid        NOT NULL REFERENCES titles (id) ON DELETE CASCADE,
    season_number integer     NOT NULL CHECK (season_number > 0),
    name          text        NOT NULL,
    air_date      date,
    episode_count integer     NOT NULL DEFAULT 0 CHECK (episode_count >= 0),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (title_id, season_number)
);

-- +goose Down
DROP TABLE seasons;
DROP TABLE titles;
