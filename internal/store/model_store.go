package store

import (
	"database/sql"
	"encoding/json"
	"errors"

	"task214-seepage/internal/model"
)

// CreateBoundaryModel persists a candidate boundary model.
func (s *Store) CreateBoundaryModel(m *model.BoundaryModel) error {
	layers, err := json.Marshal(m.Layers)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO boundary_models(id,experiment_id,name,form,layers,state,created_at)
		 VALUES(?,?,?,?,?,?,?)`,
		m.ID, m.ExperimentID, m.Name, string(m.Form), string(layers), string(m.State), fmtTime(m.CreatedAt),
	)
	return err
}

// GetBoundaryModel loads a boundary model by id.
func (s *Store) GetBoundaryModel(id string) (*model.BoundaryModel, error) {
	row := s.db.QueryRow(
		`SELECT id,experiment_id,name,form,layers,state,created_at FROM boundary_models WHERE id=?`, id)
	return scanBoundaryModel(row)
}

// ListBoundaryModels returns all models of an experiment.
func (s *Store) ListBoundaryModels(expID string) ([]*model.BoundaryModel, error) {
	rows, err := s.db.Query(
		`SELECT id,experiment_id,name,form,layers,state,created_at FROM boundary_models WHERE experiment_id=? ORDER BY created_at ASC`, expID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.BoundaryModel
	for rows.Next() {
		m, err := scanBoundaryModel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// UpdateBoundaryModelState sets the lifecycle state of a model.
func (s *Store) UpdateBoundaryModelState(id string, state model.ModelState) error {
	res, err := s.db.Exec(`UPDATE boundary_models SET state=? WHERE id=?`, string(state), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func scanBoundaryModel(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.BoundaryModel, error) {
	var (
		id, expID, name, form, state, createdAt, layersJSON sql.NullString
	)
	if err := scanner.Scan(&id, &expID, &name, &form, &layersJSON, &state, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	var layers []model.Layer
	if err := json.Unmarshal([]byte(layersJSON.String), &layers); err != nil {
		return nil, err
	}
	return &model.BoundaryModel{
		ID:           id.String,
		ExperimentID: expID.String,
		Name:         name.String,
		Form:         model.ModelForm(form.String),
		Layers:       layers,
		State:        model.ModelState(state.String),
		CreatedAt:    parseTime(createdAt.String),
	}, nil
}
