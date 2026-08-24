package boundary

import "task214-seepage/internal/model"

// CandidatePair is a conventional uniform + layered pair of starting candidates
// used when an experiment enters the inversion stage.
type CandidatePair struct {
	Uniform *model.BoundaryModel
	Layered *model.BoundaryModel
}

// SeedDefaults creates a uniform and a layered candidate for an experiment and
// marks them runnable. The uniform guess uses the geometric mean of the layered
// pair so the two candidates bracket the true medium.
func (s *Service) SeedDefaults(expID string, kInlet, kOutlet float64) (*CandidatePair, error) {
	uniformGuess := (kInlet + kOutlet) / 2
	u, err := s.CreateUniform(expID, "uniform-homogeneous", uniformGuess)
	if err != nil {
		return nil, err
	}
	l, err := s.CreateLayered(expID, "layered-inlet/outlet", kInlet, kOutlet)
	if err != nil {
		return nil, err
	}
	if err := s.SetState(u.ID, model.ModelRunnable); err != nil {
		return nil, err
	}
	u.State = model.ModelRunnable
	if err := s.SetState(l.ID, model.ModelRunnable); err != nil {
		return nil, err
	}
	l.State = model.ModelRunnable
	return &CandidatePair{Uniform: u, Layered: l}, nil
}
