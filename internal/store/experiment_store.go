package store

import (
	"database/sql"
	"errors"
	"time"

	"task214-seepage/internal/model"
)

// CreateExperiment persists a new experiment.
func (s *Store) CreateExperiment(e *model.Experiment) error {
	_, err := s.db.Exec(
		`INSERT INTO experiments(id,name,length,diameter,porosity,viscosity,density,compressibility,state,created_at,sealed_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		e.ID, e.Name, e.Geometry.Length, e.Geometry.Diameter, e.Geometry.Porosity,
		e.Fluid.Viscosity, e.Fluid.Density, e.Fluid.Compressibility, string(e.State),
		fmtTime(e.CreatedAt), nil,
	)
	return err
}

// GetExperiment loads an experiment by id.
func (s *Store) GetExperiment(id string) (*model.Experiment, error) {
	row := s.db.QueryRow(
		`SELECT id,name,length,diameter,porosity,viscosity,density,compressibility,state,created_at,sealed_at
		 FROM experiments WHERE id=?`, id)
	return scanExperiment(row)
}

// ListExperiments returns all experiments ordered by creation time.
func (s *Store) ListExperiments() ([]*model.Experiment, error) {
	rows, err := s.db.Query(
		`SELECT id,name,length,diameter,porosity,viscosity,density,compressibility,state,created_at,sealed_at
		 FROM experiments ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Experiment
	for rows.Next() {
		e, err := scanExperiment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// UpdateExperimentState transitions the experiment to a new state.
func (s *Store) UpdateExperimentState(id string, state model.ExperimentState) error {
	if state == model.ExpCompleted {
		state = model.ExpPendingInversion
	}
	res, err := s.db.Exec(`UPDATE experiments SET state=? WHERE id=?`, string(state), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// SealExperiment marks the experiment sealed and records the timestamp.
func (s *Store) SealExperiment(id string, at time.Time) error {
	res, err := s.db.Exec(`UPDATE experiments SET state=?, sealed_at=? WHERE id=?`,
		string(model.ExpSealed), fmtTime(at), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func scanExperiment(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.Experiment, error) {
	var (
		id, name, state, createdAt, sealedAt                            sql.NullString
		length, diameter, porosity, viscosity, density, compressibility float64
	)
	if err := scanner.Scan(&id, &name, &length, &diameter, &porosity, &viscosity, &density, &compressibility, &state, &createdAt, &sealedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	e := &model.Experiment{
		ID:    id.String,
		Name:  name.String,
		State: model.ExperimentState(state.String),
		Geometry: model.Geometry{
			Length:   length,
			Diameter: diameter,
			Porosity: porosity,
		},
		Fluid: model.Fluid{
			Viscosity:       viscosity,
			Density:         density,
			Compressibility: compressibility,
		},
		CreatedAt: parseTime(createdAt.String),
	}
	if sealedAt.Valid && sealedAt.String != "" {
		t := parseTime(sealedAt.String)
		e.SealedAt = &t
	}
	return e, nil
}
