CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE model_type AS ENUM ('good', 'mid', 'bad');

CREATE TABLE IF NOT EXISTS model (
    id          SERIAL      PRIMARY KEY,
    name        TEXT        NOT NULL,
    typ         model_type  DEFAULT 'mid',
    cat_uuid    UUID,
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);

INSERT INTO model(name)
VALUES
    ('test1'),
    ('test2'),
    ('test3');

CREATE TABLE IF NOT EXISTS model_info (
    model_id        INTEGER     NOT NULL PRIMARY KEY REFERENCES model(id),
    info            TEXT        NOT NULL,
    tags            TEXT[]      DEFAULT '{"available"}',
    availability    int4range   DEFAULT '[2020, 2025)',
    created_at      TIMESTAMP   NOT NULL DEFAULT NOW()
);
