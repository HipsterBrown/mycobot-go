package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServo_Structure(t *testing.T) {
	servo := &Servo{}
	assert.NotNil(t, servo)
}

func TestServo_MethodsExist(t *testing.T) {
	servo := &Servo{}
	_ = servo.ReleaseServo
	_ = servo.FocusServo
	_ = servo.IsServoEnabled
	_ = servo.GetEncoder
	_ = servo.SetEncoder
	_ = servo.GetEncoders
	_ = servo.SetEncoders
	_ = servo.GetServoData
	_ = servo.SetServoData
	_ = servo.SetServoCalibration
	_ = servo.GetJointMin
	_ = servo.GetJointMax
}
