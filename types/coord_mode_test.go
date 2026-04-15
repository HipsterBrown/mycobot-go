package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordMode_Constants(t *testing.T) {
	assert.Equal(t, CoordMode(0), CoordModeAngular)
	assert.Equal(t, CoordMode(1), CoordModeLinear)
}

func TestCoordMode_Validate(t *testing.T) {
	assert.NoError(t, CoordModeAngular.Validate())
	assert.NoError(t, CoordModeLinear.Validate())
	assert.Error(t, CoordMode(2).Validate())
}
