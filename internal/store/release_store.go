package store

import (
	"database/sql"
	"errors"

	"task214-seepage/internal/model"
)

// CreateReleaseVersion persists a published version.
func (s *Store) CreateReleaseVersion(r *model.ReleaseVersion) error {
	_, err := s.db.Exec(
		`INSERT INTO release_versions(id,experiment_id,source_model_id,source_inversion_id,permeability,interval_lo,interval_hi,notes,seq,created_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		r.ID, r.ExperimentID, r.SourceModelID, r.SourceInversionID, r.Permeability,
		r.IntervalLo, r.IntervalHi, r.Notes, r.Seq, fmtTime(r.CreatedAt),
	)
	return err
}

// GetReleaseVersion loads a release version by id.
func (s *Store) GetReleaseVersion(id string) (*model.ReleaseVersion, error) {
	row := s.db.QueryRow(
		`SELECT id,experiment_id,source_model_id,source_inversion_id,permeability,interval_lo,interval_hi,notes,seq,created_at
		 FROM release_versions WHERE id=?`, id)
	return scanRelease(row)
}

// ListReleaseVersions returns all release versions of an experiment ordered by seq.
func (s *Store) ListReleaseVersions(expID string) ([]*model.ReleaseVersion, error) {
	rows, err := s.db.Query(
		`SELECT id,experiment_id,source_model_id,source_inversion_id,permeability,interval_lo,interval_hi,notes,seq,created_at
		 FROM release_versions WHERE experiment_id=? ORDER BY seq ASC`, expID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.ReleaseVersion
	for rows.Next() {
		r, err := scanRelease(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// NextReleaseSeq returns the next sequence number for an experiment.
func (s *Store) NextReleaseSeq(expID string) (int, error) {
	var n sql.NullInt64
	err := s.db.QueryRow(`SELECT COALESCE(MAX(seq),0) FROM release_versions WHERE experiment_id=?`, expID).Scan(&n)
	if err != nil {
		return 0, err
	}
	return int(n.Int64) + 1, nil
}

func scanRelease(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.ReleaseVersion, error) {
	var (
		id, expID, modelID, invID, notes, createdAt sql.NullString
		perm, lo, hi                                  float64
		seq                                           int
	)
	if err := scanner.Scan(&id, &expID, &modelID, &invID, &perm, &lo, &hi, &notes, &seq, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &model.ReleaseVersion{
		ID:                id.String,
		ExperimentID:      expID.String,
		SourceModelID:     modelID.String,
		SourceInversionID: invID.String,
		Permeability:      perm,
		IntervalLo:        lo,
		IntervalHi:        hi,
		Notes:             notes.String,
		Seq:               seq,
		CreatedAt:         parseTime(createdAt.String),
	}, nil
}
