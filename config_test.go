package mycobot

import (
	"testing"

	"github.com/hipsterbrown/mycobot-go/types"
	"github.com/stretchr/testify/assert"
)

func TestModelConfig_MechArm270(t *testing.T) {
	config := getModelConfig(types.ModelMechArm270)

	assert.Equal(t, types.ModelMechArm270, config.Model)
	assert.Equal(t, 6, config.JointCount)
	assert.Equal(t, 115200, config.DefaultBaud)
	assert.False(t, config.UseCRC, "CRC should be off by default, matching pymycobot")
	assert.Len(t, config.JointLimits, 6)
}
