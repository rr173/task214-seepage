// Package store provides SQLite persistence for the seepage inversion service.
// All tables use a single-writer connection (SQLite embedded) and store JSON
// payloads for nested entities so that restarts recover the full state.
package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS experiments (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    length       REAL NOT NULL,
    diameter     REAL NOT NULL,
    porosity     REAL NOT NULL,
    viscosity    REAL NOT NULL,
    density      REAL NOT NULL,
    compressibility REAL NOT NULL,
    state        TEXT NOT NULL,
    created_at   TEXT NOT NULL,
    sealed_at    TEXT
);
CREATE TABLE IF NOT EXISTS sensor_sequences (
    id           TEXT PRIMARY KEY,
    experiment_id TEXT NOT NULL,
    sensor_type  TEXT NOT NULL,
    location     TEXT NOT NULL,
    x_frac       REAL NOT NULL,
    stage        TEXT NOT NULL,
    unit         TEXT NOT NULL,
    samples      TEXT NOT NULL,
    window_hash  TEXT NOT NULL UNIQUE,
    state        TEXT NOT NULL,
    created_at   TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS boundary_models (
    id           TEXT PRIMARY KEY,
    experiment_id TEXT NOT NULL,
    name         TEXT NOT NULL,
    form         TEXT NOT NULL,
    layers       TEXT NOT NULL,
    state        TEXT NOT NULL,
    created_at   TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS inversion_tasks (
    id               TEXT PRIMARY KEY,
    model_id         TEXT NOT NULL,
    experiment_id    TEXT NOT NULL,
    state            TEXT NOT NULL,
    estimated_layers TEXT NOT NULL,
    residual_norm    REAL NOT NULL,
    normalized_sse   REAL NOT NULL,
    ident_lo         REAL NOT NULL,
    ident_hi         REAL NOT NULL,
    observations_used INTEGER NOT NULL,
    created_at       TEXT NOT NULL,
    completed_at     TEXT
);
CREATE TABLE IF NOT EXISTS release_versions (
    id                 TEXT PRIMARY KEY,
    experiment_id      TEXT NOT NULL,
    source_model_id    TEXT NOT NULL,
    source_inversion_id TEXT NOT NULL,
    permeability       REAL NOT NULL,
    interval_lo        REAL NOT NULL,
    interval_hi        REAL NOT NULL,
    notes              TEXT NOT NULL,
    seq                INTEGER NOT NULL,
    created_at         TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_seq_exp ON sensor_sequences(experiment_id);
CREATE INDEX IF NOT EXISTS idx_model_exp ON boundary_models(experiment_id);
CREATE INDEX IF NOT EXISTS idx_inv_exp ON inversion_tasks(experiment_id);
CREATE INDEX IF NOT EXISTS idx_rel_exp ON release_versions(experiment_id);
`

// Store wraps the SQLite connection.
type Store struct {
	db *sql.DB
}

// Open connects to (or creates) the database at path and applies the schema.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

// DB exposes the underlying handle for advanced callers.
func (s *Store) DB() *sql.DB { return s.db }
