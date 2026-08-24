package trial

import (
	"errors"

	"task214-seepage/internal/model"
)

// ValidateGeometry checks that the column dimensions are physically sane.
func ValidateGeometry(g model.Geometry) error {
	if g.Length <= 0 {
		return errors.New("geometry: length must be positive")
	}
	if g.Diameter <= 0 {
		return errors.New("geometry: diameter must be positive")
	}
	if g.Porosity <= 0 || g.Porosity >= 1 {
		return errors.New("geometry: porosity must be in (0,1)")
	}
	return nil
}

// ValidateFluid checks that the fluid properties are positive.
func ValidateFluid(f model.Fluid) error {
	if f.Viscosity <= 0 {
		return errors.New("fluid: viscosity must be positive")
	}
	if f.Compressibility <= 0 {
		return errors.New("fluid: compressibility must be positive")
	}
	if f.Density < 0 {
		return errors.New("fluid: density must be non-negative")
	}
	return nil
}
