package mycobot

import (
	"context"

	"github.com/yourusername/mycobot-go/types"
)

// Robot is the base interface all robot models implement
type Robot interface {
	// Connection management
	Open(ctx context.Context) error
	Close() error
	IsConnected() bool

	// Core motion commands
	SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error
	GetAngles(ctx context.Context) (types.Angles, error)
	SendCoords(ctx context.Context, coord types.Coord, speed types.Speed) error
	GetCoords(ctx context.Context) (types.Coord, error)

	// Power and status
	PowerOn(ctx context.Context) error
	PowerOff(ctx context.Context) error
	IsPowerOn(ctx context.Context) (bool, error)

	// Movement queries
	IsMoving(ctx context.Context) (bool, error)
	IsInPosition(ctx context.Context, target types.Coord, tolerance float64) (bool, error)
}
