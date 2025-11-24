package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/hipsterbrown/mycobot-go/types"
)

func TestModelConfig_MyCobot280(t *testing.T) {
	config := getModelConfig(types.ModelMyCobot280)

	assert.Equal(t, types.ModelMyCobot280, config.Model)
	assert.Equal(t, 6, config.JointCount)
	assert.Equal(t, 115200, config.DefaultBaud)
	assert.True(t, config.UseCRC)
	assert.Len(t, config.JointLimits, 6)
}

func TestModelConfig_AllModels(t *testing.T) {
	models := []types.Model{
		types.ModelMyCobot280,
		types.ModelMyCobot320,
		types.ModelMechArm270,
		types.ModelMyPalletizer260,
	}

	for _, model := range models {
		config := getModelConfig(model)
		assert.NotNil(t, config)
		assert.Equal(t, model, config.Model)
		assert.Greater(t, config.JointCount, 0)
	}
}
