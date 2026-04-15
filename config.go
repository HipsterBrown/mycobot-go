package mycobot

import "github.com/hipsterbrown/mycobot-go/types"

// ModelConfig defines model-specific parameters
type ModelConfig struct {
	Model         types.Model
	JointCount    int
	JointLimits   []types.JointLimit
	UseCRC        bool
	DefaultBaud   int
	SupportedBaud []int
}

func getModelConfig(model types.Model) ModelConfig {
	return modelConfigs[model]
}

var modelConfigs = map[types.Model]ModelConfig{
	types.ModelMechArm270: {
		Model:      types.ModelMechArm270,
		JointCount: 6,
		JointLimits: []types.JointLimit{
			{MinAngle: -165, MaxAngle: 165}, // Joint 1
			{MinAngle: -165, MaxAngle: 165}, // Joint 2
			{MinAngle: -165, MaxAngle: 165}, // Joint 3
			{MinAngle: -165, MaxAngle: 165}, // Joint 4
			{MinAngle: -165, MaxAngle: 165}, // Joint 5
			{MinAngle: -175, MaxAngle: 175}, // Joint 6
		},
		UseCRC:        false,
		DefaultBaud:   115200,
		SupportedBaud: []int{115200, 1000000},
	},
}
