package mycobot

import "github.com/yourusername/mycobot-go/types"

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
	types.ModelMyCobot280: {
		Model:      types.ModelMyCobot280,
		JointCount: 6,
		JointLimits: []types.JointLimit{
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -175, MaxAngle: 175},
		},
		UseCRC:        true,
		DefaultBaud:   115200,
		SupportedBaud: []int{115200, 1000000},
	},
	types.ModelMyCobot320: {
		Model:      types.ModelMyCobot320,
		JointCount: 6,
		JointLimits: []types.JointLimit{
			{MinAngle: -170, MaxAngle: 170},
			{MinAngle: -137, MaxAngle: 137},
			{MinAngle: -150, MaxAngle: 150},
			{MinAngle: -145, MaxAngle: 145},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -180, MaxAngle: 180},
		},
		UseCRC:        true,
		DefaultBaud:   115200,
		SupportedBaud: []int{115200, 1000000},
	},
	types.ModelMechArm270: {
		Model:      types.ModelMechArm270,
		JointCount: 6,
		JointLimits: []types.JointLimit{
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -175, MaxAngle: 175},
		},
		UseCRC:        true,
		DefaultBaud:   115200,
		SupportedBaud: []int{115200, 1000000},
	},
	types.ModelMyPalletizer260: {
		Model:      types.ModelMyPalletizer260,
		JointCount: 4,
		JointLimits: []types.JointLimit{
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -90, MaxAngle: 90},
			{MinAngle: -90, MaxAngle: 90},
			{MinAngle: -165, MaxAngle: 165},
		},
		UseCRC:        true,
		DefaultBaud:   115200,
		SupportedBaud: []int{115200, 1000000},
	},
}
