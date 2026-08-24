package store

import (
	"database/sql"
	"encoding/json"
	"errors"

	"task214-seepage/internal/model"
)

// CreateSensorSequence inserts a sequence, enforcing idempotency by window hash.
// If an identical sequence already exists it is returned together with
// model.ErrDuplicateSequence and nothing is written.
func (s *Store) CreateSensorSequence(seq *model.SensorSequence) (*model.SensorSequence, error) {
	if existing, err := s.GetSensorSequenceByHash(seq.WindowHash); err == nil {
		return existing, model.ErrDuplicateSequence
	}
	samples, err := json.Marshal(seq.Samples)
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec(
		`INSERT INTO sensor_sequences(id,experiment_id,sensor_type,location,x_frac,stage,unit,samples,window_hash,state,created_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		seq.ID, seq.ExperimentID, string(seq.SensorType), seq.Location, seq.XFrac, seq.Stage,
		seq.Unit, string(samples), seq.WindowHash, string(seq.State), fmtTime(seq.CreatedAt),
	)
	if err != nil {
		return nil, err
	}
	return seq, nil
}

// GetSensorSequence loads a sequence by id.
func (s *Store) GetSensorSequence(id string) (*model.SensorSequence, error) {
	row := s.db.QueryRow(
		`SELECT id,experiment_id,sensor_type,location,x_frac,stage,unit,samples,window_hash,state,created_at
		 FROM sensor_sequences WHERE id=?`, id)
	return scanSensor(row)
}

// GetSensorSequenceByHash finds a previously stored sequence by its window hash.
func (s *Store) GetSensorSequenceByHash(hash string) (*model.SensorSequence, error) {
	row := s.db.QueryRow(
		`SELECT id,experiment_id,sensor_type,location,x_frac,stage,unit,samples,window_hash,state,created_at
		 FROM sensor_sequences WHERE window_hash=?`, hash)
	return scanSensor(row)
}

// ListSensorSequences returns all sequences of an experiment.
func (s *Store) ListSensorSequences(expID string) ([]*model.SensorSequence, error) {
	rows, err := s.db.Query(
		`SELECT id,experiment_id,sensor_type,location,x_frac,stage,unit,samples,window_hash,state,created_at
		 FROM sensor_sequences WHERE experiment_id=? ORDER BY created_at ASC`, expID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.SensorSequence
	for rows.Next() {
		seq, err := scanSensor(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, seq)
	}
	return out, rows.Err()
}

// UpdateSensorState sets the validation state of a sequence.
func (s *Store) UpdateSensorState(id string, state model.SequenceState) error {
	res, err := s.db.Exec(`UPDATE sensor_sequences SET state=? WHERE id=?`, string(state), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func scanSensor(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.SensorSequence, error) {
	var (
		id, expID, sensorType, location, stage, unit, state, createdAt, hash sql.NullString
		xFrac                                                                   float64
		samplesJSON                                                            string
	)
	if err := scanner.Scan(&id, &expID, &sensorType, &location, &xFrac, &stage, &unit, &samplesJSON, &hash, &state, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	var samples []model.Sample
	if err := json.Unmarshal([]byte(samplesJSON), &samples); err != nil {
		return nil, err
	}
	return &model.SensorSequence{
		ID:           id.String,
		ExperimentID: expID.String,
		SensorType:   model.SensorType(sensorType.String),
		Location:     location.String,
		XFrac:        xFrac,
		Stage:        stage.String,
		Unit:         unit.String,
		Samples:      samples,
		WindowHash:   hash.String,
		State:        model.SequenceState(state.String),
		CreatedAt:    parseTime(createdAt.String),
	}, nil
}
