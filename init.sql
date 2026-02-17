\connect scratchdb;

CREATE TABLE IF NOT EXISTS scratch (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    lastupdated TIMESTAMP DEFAULT NOW(),
    password TEXT NOT NULL
);
