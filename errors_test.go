package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandardErrors_Defined(t *testing.T) {
	// Just verify errors are defined
	assert.NotNil(t, ErrRobotClosed)
	assert.NotNil(t, ErrNotConnected)
}
