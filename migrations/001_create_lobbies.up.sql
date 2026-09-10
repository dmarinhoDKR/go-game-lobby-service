CREATE TABLE lobbies (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    max_players BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT lobbies_name_not_blank
        CHECK (name ~ '[^[:space:]]'),

    CONSTRAINT lobbies_status_valid
        CHECK (status IN ('waiting', 'started', 'closed')),

    CONSTRAINT lobbies_max_players_minimum
        CHECK (max_players >= 2)
);
