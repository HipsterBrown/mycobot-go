package mycobot

import "github.com/hipsterbrown/mycobot-go/internal/robot"

// IO provides Atom IO operations (end-effector head, 0x60 range)
type IO struct {
	robot *robot.Base
}
