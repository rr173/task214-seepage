package store

import (
	"database/sql"
	"encoding/json"
	"errors"

	"task214-seepage/internal/model"
)

// CreateInversionTask persists a newly created inversion task.
func (s *Store) CreateInversionTask(t *model.InversionTask) error {
	layers, err := json.Marshal(t.EstimatedLayers)
	if err != nil {
		return err
	}
	completed := ""
	if t.CompletedAt != nil {
		completed = fmtTime(*t.CompletedAt)
	}
	_, err = s.db.Exec(
		`INSERT INTO inversion_tasks(id,model_id,experiment_id,state,estimated_layers,residual_norm,normalized_sse,ident_lo,ident_hi,observations_used,created_at,completed_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.ModelID, t.ExperimentID, string(t.State), string(layers),
		t.ResidualNorm, t.NormalizedSSE, t.IdentifiabilityLo, t.IdentifiabilityHi,
		t.ObservationsUsed, fmtTime(t.CreatedAt), completed,
	)
	return err
}

// UpdateInversionTask writes the result of a completed inversion.
func (s *Store) UpdateInversionTask(t *model.InversionTask) error {
	layers, err := json.Marshal(t.EstimatedLayers)
	if err != nil {
		return err
	}
	completed := ""
	if t.CompletedAt != nil {
		completed = fmtTime(*t.CompletedAt)
	}
	res, err := s.db.Exec(
		`UPDATE inversion_tasks SET state=?,estimated_layers=?,residual_norm=?,normalized_sse=?,ident_lo=?,ident_hi=?,observations_used=?,completed_at=?
		 WHERE id=?`,
		string(t.State), string(layers), t.ResidualNorm, t.NormalizedSSE,
		t.IdentifiabilityLo, t.IdentifiabilityHi, t.ObservationsUsed, completed, t.ID,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// GetInversionTask loads an inversion task by id.
func (s *Store) GetInversionTask(id string) (*model.InversionTask, error) {
	row := s.db.QueryRow(
		`SELECT id,model_id,experiment_id,state,estimated_layers,residual_norm,normalized_sse,ident_lo,ident_hi,observations_used,created_at,completed_at
		 FROM inversion_tasks WHERE id=?`, id)
	return scanInversion(row)
}

// ListInversionTasks returns all inversion tasks of an experiment.
func (s *Store) ListInversionTasks(expID string) ([]*model.InversionTask, error) {
	rows, err := s.db.Query(
		`SELECT id,model_id,experiment_id,state,estimated_layers,residual_norm,normalized_sse,ident_lo,ident_hi,observations_used,created_at,completed_at
		 FROM inversion_tasks WHERE experiment_id=? ORDER BY created_at ASC`, expID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.InversionTask
	for rows.Next() {
		t, err := scanInversion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func scanInversion(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.InversionTask, error) {
	var (
		id, modelID, expID, state, createdAt, completedAt, layersJSON sql.NullString
		residual, normSSE, lo, hi                                     float64
		obsUsed                                                       int
	)
	if err := scanner.Scan(&id, &modelID, &expID, &state, &layersJSON, &residual, &normSSE, &lo, &hi, &obsUsed, &createdAt, &completedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	var layers []model.Layer
	if err := json.Unmarshal([]byte(layersJSON.String), &layers); err != nil {
		return nil, err
	}
	t := &model.InversionTask{
		ID:                id.String,
		ModelID:           modelID.String,
		ExperimentID:      expID.String,
		State:             model.InversionState(state.String),
		EstimatedLayers:   layers,
		ResidualNorm:      residual,
		NormalizedSSE:     normSSE,
		IdentifiabilityLo: lo,
		IdentifiabilityHi: hi,
		ObservationsUsed:  obsUsed,
		CreatedAt:         parseTime(createdAt.String),
	}
	if completedAt.Valid && completedAt.String != "" {
		t.CompletedAt = parseNullableTime(completedAt.String)
	}
	return t, nil
}
