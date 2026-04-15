package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirection_Constants(t *testing.T) {
	assert.Equal(t, Direction(0), DirNegative)
	assert.Equal(t, Direction(1), DirPositive)
}

func TestDirection_Validate(t *testing.T) {
	assert.NoError(t, DirNegative.Validate())
	assert.NoError(t, DirPositive.Validate())
	assert.Error(t, Direction(2).Validate())
}
